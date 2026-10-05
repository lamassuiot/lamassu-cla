#!/usr/bin/env bash
# Private-URL scan (ADR-0003): fails when a URL's host contains an internal, intranet, or corp label.
# Usage: scripts/check-private-urls.sh [repository-root]
# Exit codes: 0 clean, 1 findings, 2 invalid allowlist.
set -euo pipefail

cd "${1:-.}"

allowlist_name=".github/private-url-allowlist.txt"
pattern='[A-Za-z][A-Za-z0-9+.-]*://([^/@[:space:]]+@)?([A-Za-z0-9-]+\.)*(internal|intranet|corp)([^A-Za-z0-9-]|$)'

allowed=()
if [[ -f "$allowlist_name" ]]; then
  previous=""
  line_number=0
  while IFS= read -r line || [[ -n "$line" ]]; do
    line_number=$((line_number + 1))
    if [[ -z "$line" || "$line" == \#* ]]; then
      previous="$line"
      continue
    fi
    if [[ "$previous" != \#* ]]; then
      echo "::error file=${allowlist_name},line=${line_number}::Each allowlist entry needs a justification comment on the line directly above it."
      exit 2
    fi
    allowed+=("$line")
    previous="$line"
  done < "$allowlist_name"
fi

targets=()
while IFS= read -r -d '' file; do
  [[ "$file" == "$allowlist_name" ]] || targets+=("$file")
done < <(git ls-files -z --cached --others --exclude-standard)

matches=""
if ((${#targets[@]})); then
  matches=$(printf '%s\0' "${targets[@]}" | xargs -0 -r grep -InEs -- "$pattern" || true)
fi

if [[ -n "$matches" && ${#allowed[@]} -gt 0 ]]; then
  filters=()
  for entry in "${allowed[@]}"; do
    filters+=(-e "$entry")
  done
  matches=$(grep -vF "${filters[@]}" <<<"$matches" || true)
fi

if [[ -n "$matches" ]]; then
  echo "::error::Possible private hostname in a URL. Remove it, or add a justified entry to ${allowlist_name}."
  echo "$matches"
  exit 1
fi

echo "No private hostnames found in URLs."
