#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <snapshot-name-or-bundle-image>" >&2
    exit 2
fi

INPUT="$1"

get_bundle_image_from_snapshot() {
    local snapshot_name="$1"
    echo "Getting VCF Migration Operator bundle image from snapshot: $snapshot_name" >&2
    oc describe snapshot "$snapshot_name" |
        awk '/Container Image:/ && /bundle/ {print $NF; exit}'
}

if [[ "$INPUT" == *"@"* || ( "$INPUT" == */* && ( "${INPUT%%/*}" == *.* || "${INPUT%%/*}" == *:* ) ) ]]; then
    BUNDLE_IMAGE="$INPUT"
    echo "Using provided bundle image: $BUNDLE_IMAGE"
else
    echo "Input appears to be a snapshot. Extracting bundle image..."
    BUNDLE_IMAGE=$(get_bundle_image_from_snapshot "$INPUT")
    if [[ -z "$BUNDLE_IMAGE" ]]; then
        echo "Error: Could not find a bundle image in snapshot $INPUT" >&2
        exit 1
    fi
    echo "Found bundle image: $BUNDLE_IMAGE"
fi

if ! command -v opm >/dev/null 2>&1; then
    echo "Error: opm must be installed and available on PATH" >&2
    exit 1
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

found_catalog=false
for dir in fbc-v*-*; do
    [[ -d "$dir" ]] || continue
    if [[ ! "$dir" =~ ^fbc-v([0-9]+)-([0-9]+)$ ]]; then
        echo "Warning: Skipping unrecognized catalog directory: $dir" >&2
        continue
    fi

    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    catalog="$dir/catalog/vcf-migration-operator/catalog.yaml"
    if [[ ! -f "$catalog" ]]; then
        echo "Warning: catalog not found: $catalog" >&2
        continue
    fi
    found_catalog=true

    echo "Processing $catalog"
    rendered="$TMP_DIR/${dir}.rendered.yaml"
    cleaned="$TMP_DIR/${dir}.cleaned.yaml"
    combined="$TMP_DIR/${dir}.catalog.yaml"

    if (( major > 4 || (major == 4 && minor >= 17) )); then
        opm render --output=yaml --migrate-level bundle-object-to-csv-metadata "$BUNDLE_IMAGE" > "$rendered"
    else
        opm render --output=yaml "$BUNDLE_IMAGE" > "$rendered"
    fi

    # Rewrite the rendered bundle pullspecs from the internal Quay build image
    # to the official registry.redhat.io location, preserving the digest.
    sed -i -E 's|^(image: )quay.io/redhat-user-workloads[^@]*vcf-migration-operator-bundle[^@]*@sha256:([a-f0-9]{64})|\1registry.redhat.io/vcf-migration-operator-tech-preview/vcf-migration-operator-bundle@sha256:\2|' "$rendered"
    sed -i -E 's|^([[:space:]]*- image: )quay.io/redhat-user-workloads[^@]*vcf-migration-operator-bundle[^@]*@sha256:([a-f0-9]{64})|\1registry.redhat.io/vcf-migration-operator-tech-preview/vcf-migration-operator-bundle@sha256:\2|' "$rendered"

    if ! grep -q '^schema: olm\.bundle$' "$rendered"; then
        echo "Error: opm render did not produce an olm.bundle record for $BUNDLE_IMAGE" >&2
        exit 1
    fi

    # Remove the scaffold's placeholder bundle document before adding the rendered bundle.
    awk 'BEGIN { RS = "---[[:space:]]*\n"; ORS = "" }
        NR > 1 && index($0, "# TODO: Replace this placeholder with the actual rendered bundle.") == 0 {
            printf "---\n%s", $0
        }' "$catalog" > "$cleaned"

    # opm render may emit the bundle name before the schema field.
    bundle_name=$(awk '/^name:/ { sub(/^name:[[:space:]]*/, ""); print; exit }' "$rendered")
    if [[ -z "$bundle_name" ]]; then
        echo "Error: Could not determine the bundle name rendered from $BUNDLE_IMAGE" >&2
        exit 1
    fi
    if [[ "$bundle_name" != vcf-migration-operator.* ]]; then
        echo "Error: Rendered bundle does not appear to belong to vcf-migration-operator: $bundle_name" >&2
        exit 1
    fi

    if awk -v name="$bundle_name" '$0 == "name: " name { found = 1 } END { exit !found }' "$cleaned"; then
        echo "Bundle $bundle_name is already present in $catalog; skipping"
        continue
    fi

    awk -v name="$bundle_name" \
        '$0 == "  - name: vcf-migration-operator.v0.0.1" { $0 = "  - name: " name } { print }' \
        "$cleaned" > "$combined"
    cat "$rendered" >> "$combined"
    chmod --reference="$catalog" "$combined"
    mv "$combined" "$catalog"
    echo "Added $bundle_name to $catalog"
done

if [[ "$found_catalog" != true ]]; then
    echo "Error: No VCF catalog.yaml files found under fbc-v* directories" >&2
    exit 1
fi

echo "Done."
