#!/usr/bin/env bash
# Tests for the .biome/plugins/no-component-copy.grit plugin as configured in biome.json
# (F-05 AC-05-9, F05-T9): component source must take all user-visible text through props.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
biome="$root/node_modules/.bin/biome"
pass=0
fail=0

# lint_case NAME PATH EXPECTED CONTENT: EXPECTED is "copy" (plugin must report) or "clean".
lint_case() {
  local name="$1" path="$2" expected="$3" content="$4" dir output status=0
  dir="$(mktemp -d)"
  git -C "$dir" init -q
  cp -R "$root/.biome" "$dir/.biome"
  # shellcheck disable=SC2016 # The JavaScript program is intentionally single-quoted.
  node -e 'const c = require(process.argv[1]); delete c.$schema; console.log(JSON.stringify(c));' \
    "$root/biome.json" >"$dir/biome.json"
  mkdir -p "$dir/$(dirname "$path")"
  printf '%s\n' "$content" >"$dir/$path"
  output="$(cd "$dir" && "$biome" lint --max-diagnostics=none "$path" 2>&1)" || status=$?
  rm -rf "$dir"
  local reported=no
  [[ "$output" == *"through props"* ]] && reported=yes
  if [[ "$expected" == copy && "$reported" == yes && $status -ne 0 ]] ||
    [[ "$expected" == clean && "$reported" == no ]]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "FAIL: $name (expected $expected, exit $status)"
  fi
}

component=packages/design-system/src/components/demo/demo.tsx

lint_case 'literal JSX text' "$component" copy 'export const A = () => <button type="button">Save</button>;'
lint_case 'string literal child' "$component" copy "export const A = () => <p>{'Saved'}</p>;"
# shellcheck disable=SC2016 # The backticks are a TypeScript template literal, not a shell expansion.
lint_case 'template literal child' "$component" copy 'export const A = () => <p>{`Saved`}</p>;'
lint_case 'literal aria-label' "$component" copy 'export const A = () => <span aria-label="Close" />;'
lint_case 'literal alt' "$component" copy 'export const A = () => <img alt="Logo" src="/a.png" />;'
lint_case 'literal placeholder' "$component" copy 'export const A = () => <input placeholder="Search" />;'
lint_case 'literal title' "$component" copy 'export const A = () => <abbr title="Contributor License Agreement" />;'
lint_case 'text from props' "$component" clean 'export const A = ({ label }: { label: string }) => <button type="button" aria-label={label}>{label}</button>;'
lint_case 'non-text attributes' "$component" clean 'export const A = () => <input type="text" id="x" aria-describedby="x-hint" />;'
lint_case 'whitespace between elements' "$component" clean 'export const A = ({ a, b }: { a: string; b: string }) => (
  <p>
    {a}
    <span>{b}</span>
  </p>
);'
lint_case 'stories may contain copy' packages/design-system/src/components/demo/demo.stories.tsx clean 'export const A = () => <button type="button">Save</button>;'
lint_case 'tests may contain copy' packages/design-system/src/components/demo/demo.browser.test.tsx clean 'export const A = () => <button type="button">Save</button>;'
# F-06 AC-06-6: portal screens take their text from the copy module.
lint_case 'portal screen with literal text' apps/web/src/screens/demo.tsx copy 'export const A = () => <h1>Agreements</h1>;'
lint_case 'portal screen with literal aria-label' apps/web/src/shell/demo.tsx copy 'export const A = () => <nav aria-label="Main" />;'
lint_case 'portal copy module' apps/web/src/copy/demo.tsx clean 'export const A = () => <h1>Agreements</h1>;'
lint_case 'portal tests may contain copy' apps/web/src/demo.browser.test.tsx clean 'export const A = () => <h1>Agreements</h1>;'

echo "check-component-copy tests: $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
