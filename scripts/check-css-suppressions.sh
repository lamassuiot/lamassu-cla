#!/usr/bin/env bash
# Rejects Biome suppression comments in CSS, so the token-only check cannot be switched off
# (F-04 AC-04-6). An exception is a token or specification change, not a suppression.
set -euo pipefail

root="${1:-.}"
cd "$root"

mapfile -t files < <(git ls-files --cached --others --exclude-standard -- '*.css')
[[ ${#files[@]} -eq 0 ]] && exit 0

if hits=$(grep -nHE 'biome-ignore' -- "${files[@]}"); then
  while IFS= read -r hit; do
    echo "${hit%%:*}:$(cut -d: -f2 <<<"$hit"): Biome suppressions are not allowed in CSS; use a design token."
  done <<<"$hits"
  exit 1
fi
