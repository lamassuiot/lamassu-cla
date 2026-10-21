#!/usr/bin/env bash
# Regenerates the API client from specs/api/openapi.yaml and fails if the committed output differs
# (AC-02-7, decision A-22). Run from a clean tree: on drift, the regenerated files are left in place.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
generated=packages/api-client/src/generated
cd "$root"

npm run --silent generate -w @lamassu-cla/api-client >/dev/null

# Compare with the index, so a staged regeneration passes before it is committed.
if ! git diff --quiet -- "$generated" || [[ -n "$(git ls-files --others --exclude-standard -- "$generated")" ]]; then
  git status --short -- "$generated"
  git --no-pager diff --stat -- "$generated"
  echo "The generated API client is out of date with specs/api/openapi.yaml."
  echo "Run 'npm run generate -w @lamassu-cla/api-client' and commit $generated."
  exit 1
fi
echo "Generated API client matches specs/api/openapi.yaml."
