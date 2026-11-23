#!/bin/sh
# Verify that a Konflux bundle references an operator image from its snapshot.
# Requires oc login to Konflux (current namespace), yq, and podman registry access.
set -eu

usage() {
    echo "Usage: $0 [-v <main|version-with-dashes>] [-s <snapshot-name>]" >&2
    exit 1
}

fail() {
    echo "Error: $*" >&2
    exit 1
}

VERSION=main
SNAPSHOT=
while getopts "v:s:" opt; do
    case "$opt" in
        v) VERSION=$OPTARG ;;
        s) SNAPSHOT=$OPTARG ;;
        *) usage ;;
    esac
done
shift $((OPTIND - 1))
[ "$#" -eq 0 ] || usage
if [ "$VERSION" != main ]; then
    printf '%s\n' "$VERSION" | grep -Eq '^v[0-9]+(-[[:alnum:]]+)+$' ||
        fail "Version must be main or dash-separated (e.g. v0-1)."
fi

for tool in oc yq podman; do
    command -v "$tool" >/dev/null 2>&1 || fail "Required command not found: $tool"
done

if [ -z "$SNAPSHOT" ]; then
    echo "Fetching snapshots for vcf-migration-operator-$VERSION..."
    SNAPSHOTS=$(oc get snapshots --sort-by .metadata.creationTimestamp \
        -l "pac.test.appstudio.openshift.io/event-type=push,appstudio.openshift.io/application=vcf-migration-operator-$VERSION" \
        --no-headers -o custom-columns=":metadata.name")
    [ -n "$SNAPSHOTS" ] || fail "No snapshots found. Check version, Konflux login, and current namespace."
    printf '%s\n' "$SNAPSHOTS" | nl -ba
    printf 'Enter snapshot number: '
    read -r SELECTION || fail "Invalid selection"
    case "$SELECTION" in
        ''|*[!0-9]*) fail "Invalid selection" ;;
    esac
    SNAPSHOT=$(printf '%s\n' "$SNAPSHOTS" | awk -v selection="$SELECTION" 'NR == selection')
    [ -n "$SNAPSHOT" ] || fail "Invalid selection"
fi

WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/check-snapshots.XXXXXX")
CONTAINER=
cleanup() {
    if [ -n "$CONTAINER" ]; then
        podman rm "$CONTAINER" >/dev/null || echo "Warning: could not remove container $CONTAINER" >&2
    fi
    rm -rf "$WORK_DIR"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "Using snapshot: $SNAPSHOT"
oc get snapshot "$SNAPSHOT" -o yaml > "$WORK_DIR/snapshot.yaml"
yq -r '.spec.components[].containerImage' "$WORK_DIR/snapshot.yaml" > "$WORK_DIR/images"
echo 'Container images in snapshot:'
cat "$WORK_DIR/images"
yq -r '.spec.components[] | select(.name | test("bundle")) | .containerImage' \
    "$WORK_DIR/snapshot.yaml" > "$WORK_DIR/bundles"
[ "$(wc -l < "$WORK_DIR/bundles")" -eq 1 ] || fail "Expected exactly one bundle image in snapshot."
BUNDLE_IMAGE=$(head -n 1 "$WORK_DIR/bundles")
[ -n "$BUNDLE_IMAGE" ] && [ "$BUNDLE_IMAGE" != null ] || fail "Bundle image is missing."

echo "Pulling bundle image: $BUNDLE_IMAGE"
podman pull -- "$BUNDLE_IMAGE"
# Bundle is FROM scratch: copy manifests without starting a container or unpacking layers.
CONTAINER=$(podman create --entrypoint /unused -- "$BUNDLE_IMAGE")
mkdir "$WORK_DIR/manifests"
podman cp "$CONTAINER:/manifests/." "$WORK_DIR/manifests/"
set -- "$WORK_DIR"/manifests/*.clusterserviceversion.yaml
[ "$#" -eq 1 ] && [ -f "$1" ] || fail "Expected exactly one .clusterserviceversion.yaml in bundle."
yq -r '.spec.install.spec.deployments[].spec.template.spec.containers[] | select(.name == "manager") | .image' \
    "$1" > "$WORK_DIR/manager-images"
[ "$(wc -l < "$WORK_DIR/manager-images")" -eq 1 ] || fail "Expected exactly one manager image in CSV."
INTERNAL_IMAGE=$(head -n 1 "$WORK_DIR/manager-images")
echo "Manager image referenced in CSV: $INTERNAL_IMAGE"
printf '%s\n' "$INTERNAL_IMAGE" | grep -Eq '^.+@sha256:[a-f0-9]{64}$' ||
    fail "Manager image must contain a valid SHA-256 digest."
INTERNAL_SHA=${INTERNAL_IMAGE##*@}

while IFS= read -r image; do
    if [ "${image##*@}" = "$INTERNAL_SHA" ]; then
        echo 'Match found: manager image SHA is present in snapshot images.'
        exit 0
    fi
done < "$WORK_DIR/images"
fail "No match: manager image SHA is NOT present in snapshot images."
