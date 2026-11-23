#!/usr/bin/env python3
"""Run: python3 hack/test-check-snapshots.py (requires yq). No cluster needed."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).with_name("check-snapshots.sh")
SHA = "a" * 64
OTHER_SHA = "b" * 64


@unittest.skipUnless(shutil.which("yq"), "requires yq")
class CheckSnapshotsTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.env = dict(os.environ, PATH=f"{self.root}:{os.environ['PATH']}",
                        FIXTURES=str(self.root), TMPDIR=str(self.root))
        self.images = [
            {"name": "vcf-migration-operator-main", "containerImage": "quay.io/operator@sha256:" + SHA},
            {"name": "vcf-migration-operator-bundle-main", "containerImage": "quay.io/bundle@sha256:" + OTHER_SHA},
        ]
        self.manager_image = "registry.redhat.io/operator@sha256:" + SHA
        self.containers = None
        self.fake("oc", '''#!/bin/sh
set -eu
case "$*" in
  'get snapshots '*)
    case "$*" in
      *appstudio.openshift.io/application=vcf-migration-operator-main*) ;;
      *) exit 9 ;;
    esac
    printf '%s' "${SNAPSHOT_LIST-snapshot-one}" ;;
  'get snapshot snapshot-one -o yaml') cat "$FIXTURES/snapshot.json" ;;
  *) exit 9 ;;
esac
''')
        self.fake("podman", '''#!/bin/sh
set -eu
case "$1" in
  pull) exit "${PULL_STATUS-0}" ;;
  create) echo test-container ;;
  cp)
    [ "$2" = test-container:/manifests/. ] || exit 9
    cp "$FIXTURES/csv.json" "$3/test.clusterserviceversion.yaml"
    exit "${COPY_STATUS-0}" ;;
  rm) touch "$FIXTURES/removed" ;;
  *) exit 9 ;;
esac
''')

    def fake(self, name, text):
        path = self.root / name
        path.write_text(text)
        path.chmod(0o755)

    def run_check(self, *args, selection=""):
        (self.root / "snapshot.json").write_text(json.dumps({"spec": {"components": self.images}}))
        containers = self.containers
        if containers is None:
            containers = [{"name": "sidecar", "image": "ignored:latest"},
                          {"name": "manager", "image": self.manager_image}]
        csv = {"spec": {"install": {"spec": {"deployments": [
            {"spec": {"template": {"spec": {"containers": containers}}}}
        ]}}}}
        (self.root / "csv.json").write_text(json.dumps(csv))
        return subprocess.run(["sh", str(SCRIPT), *args], input=selection,
                              env=self.env, capture_output=True, text=True)

    def assert_clean(self):
        self.assertEqual(list(self.root.glob("check-snapshots.*")), [])
        self.assertTrue((self.root / "removed").exists())

    def test_explicit_snapshot_matches_digest_despite_registry_rewrite(self):
        result = self.run_check("-s", "snapshot-one")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("Match found", result.stdout)
        self.assert_clean()

    def test_mismatched_digest_fails(self):
        self.manager_image = "quay.io/operator@sha256:" + "c" * 64
        result = self.run_check("-s", "snapshot-one")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("No match", result.stderr)
        self.assert_clean()

    def test_interactive_selection_uses_main_application(self):
        result = self.run_check(selection="1\n")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_invalid_selections_fail_before_pulling(self):
        for selection in ("0\n", "2\n", "1p\n", "\n"):
            with self.subTest(selection=selection):
                result = self.run_check(selection=selection)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Invalid selection", result.stderr)

    def test_empty_snapshot_list_fails(self):
        self.env["SNAPSHOT_LIST"] = ""
        result = self.run_check()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("No snapshots", result.stderr)

    def test_invalid_version_or_extra_argument_fails(self):
        for args in (("-v", "v0.1"), ("-v", "bogus"), ("extra",)):
            with self.subTest(args=args):
                self.assertNotEqual(self.run_check(*args).returncode, 0)

    def test_missing_or_ambiguous_bundle_fails(self):
        for images in (self.images[:1], self.images + [self.images[1]]):
            with self.subTest(images=images):
                self.images = images
                result = self.run_check("-s", "snapshot-one")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("exactly one bundle", result.stderr)

    def test_missing_or_ambiguous_manager_fails(self):
        for containers in ([], [{"name": "manager", "image": self.manager_image}] * 2):
            with self.subTest(containers=containers):
                self.containers = containers
                result = self.run_check("-s", "snapshot-one")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("exactly one manager", result.stderr)
                self.assert_clean()

    def test_tag_or_malformed_digest_fails(self):
        for image in ("quay.io/operator:latest", "quay.io/operator@sha256:abc"):
            with self.subTest(image=image):
                self.manager_image = image
                result = self.run_check("-s", "snapshot-one")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("valid SHA-256", result.stderr)

    def test_external_failure_propagates_and_cleans_up(self):
        for variable in ("PULL_STATUS", "COPY_STATUS"):
            with self.subTest(variable=variable):
                self.env[variable] = "7"
                result = self.run_check("-s", "snapshot-one")
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(list(self.root.glob("check-snapshots.*")), [])
                if variable == "COPY_STATUS":
                    self.assertTrue((self.root / "removed").exists())
                del self.env[variable]


if __name__ == "__main__":
    unittest.main()
