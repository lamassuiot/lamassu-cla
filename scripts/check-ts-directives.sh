#!/usr/bin/env bash
# Rejects TypeScript comment directives Biome cannot see: triple-slash references, @ts-nocheck,
# and @ts-expect-error without a description (formerly typescript-eslint rules, ADR-0011).
set -euo pipefail

root="${1:-.}"
cd "$root"

mapfile -t files < <(git ls-files --cached --others --exclude-standard -- '*.ts' '*.tsx' '*.mts' '*.cts')
[[ ${#files[@]} -eq 0 ]] && exit 0

status=0
report() {
  local pattern="$1" message="$2" hits
  if hits=$(grep -nHE "$pattern" -- "${files[@]}"); then
    while IFS= read -r hit; do
      echo "${hit%%:*}:$(cut -d: -f2 <<<"$hit"): $message"
    done <<<"$hits"
    status=1
  fi
}

report '^[[:space:]]*///[[:space:]]*<reference[[:space:]]' 'triple-slash reference; use an import or tsconfig "types" instead.'
report '@ts-nocheck' '@ts-nocheck disables type checking for the whole file.'
report '@ts-expect-error[[:space:]]*(:[[:space:]]*)?(\*/[[:space:]]*)?$' '@ts-expect-error needs a description of why the error is expected.'

exit "$status"
