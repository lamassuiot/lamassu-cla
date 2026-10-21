#!/usr/bin/env bash
# Tests for scripts/check-design-system-components.sh.
set -euo pipefail

script="$(cd "$(dirname "$0")" && pwd)/check-design-system-components.sh"
pass=0
fail=0

layered='@layer components {
  .a {
    color: var(--cla-color-text-primary);
  }
}'

# run_case NAME EXPECTED_EXIT [FILE CONTENT]...; a complete "button" component is the baseline.
run_case() {
  local name="$1" expected="$2" dir actual
  shift 2
  dir="$(mktemp -d)"
  mkdir -p "$dir/src/components/button"
  printf "export { Button } from './components/button/button.tsx';\n" >"$dir/src/index.ts"
  printf 'export const Button = 1;\n' >"$dir/src/components/button/button.tsx"
  printf 'export default {};\n' >"$dir/src/components/button/button.stories.tsx"
  printf '%s\n' "$layered" >"$dir/src/components/button/button.module.css"
  while [[ $# -gt 0 ]]; do
    if [[ "$2" == DELETE ]]; then rm -f "$dir/$1"; else printf '%s\n' "$2" >"$dir/$1"; fi
    shift 2
  done
  if "$script" "$dir" >/dev/null 2>&1; then actual=0; else actual=1; fi
  rm -rf "$dir"
  if [[ "$actual" == "$expected" ]]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "FAIL: $name (expected exit $expected, got $actual)"
  fi
}

run_case 'complete component' 0
run_case 'missing story' 1 src/components/button/button.stories.tsx DELETE
run_case 'missing component file' 1 src/components/button/button.tsx DELETE
run_case 'not exported' 1 src/index.ts 'export {};'
run_case 'rule outside the layer' 1 src/components/button/button.module.css ".b { color: inherit; }
$layered"
run_case 'rule after the layer' 1 src/components/button/button.module.css "$layered
.b { color: inherit; }"
run_case 'wrong layer' 1 src/components/button/button.module.css '@layer base { .a { margin: 0; } }'
run_case 'leading comment is allowed' 0 src/components/button/button.module.css "/* Button. */
$layered"

echo "check-design-system-components tests: $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
