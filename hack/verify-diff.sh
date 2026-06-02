#!/bin/bash

set -eux

# Check for untracked files (excluding .idea and other common IDE directories)
FILE_DIFF="$(git ls-files -o --exclude-standard --exclude='.idea/**' --exclude='*.swp' --exclude='*.swo')"

if [ "$FILE_DIFF" != "" ]; then
  echo "Found untracked files:"
  echo "$FILE_DIFF"
  echo ""
  echo "These files should either be:"
  echo "1. Added to git (if they're part of bundle generation)"
  echo "2. Added to .gitignore (if they're build artifacts)"
  exit 1
fi

# Check for modified files
if ! git diff --exit-code; then
  echo ""
  echo "ERROR: Found uncommitted changes in tracked files."
  echo "This usually means 'make bundle' was not run before committing."
  echo ""
  echo "To fix this:"
  echo "  1. Run 'make bundle'"
  echo "  2. Commit the changes"
  exit 1
fi

echo "✓ No uncommitted changes found - bundle is up to date!"