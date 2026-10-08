#!/usr/bin/env bash
# Tests for the token-only CSS check (F-04 AC-04-6): the .biome/plugins/no-raw-css-values.grit
# plugin as configured in biome.json, and scripts/check-css-suppressions.sh.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
biome="$root/node_modules/.bin/biome"
suppressions="$root/scripts/check-css-suppressions.sh"
pass=0
fail=0

new_repo() {
  local dir
  dir="$(mktemp -d)"
  git -C "$dir" init -q
  cp -R "$root/.biome" "$dir/.biome"
  # shellcheck disable=SC2016 # The JavaScript program is intentionally single-quoted.
  node -e 'const c = require(process.argv[1]); delete c.$schema; console.log(JSON.stringify(c));' \
    "$root/biome.json" >"$dir/biome.json"
  echo "$dir"
}

record() {
  local name="$1" ok="$2" detail="$3"
  if [[ "$ok" == yes ]]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "FAIL: $name ($detail)"
  fi
}

# lint_case NAME PATH EXPECTED CSS: EXPECTED is a diagnostic message prefix, or "clean".
lint_case() {
  local name="$1" path="$2" expected="$3" content="$4" dir output status=0
  dir="$(new_repo)"
  mkdir -p "$dir/$(dirname "$path")"
  printf '%s\n' "$content" >"$dir/$path"
  output="$(cd "$dir" && "$biome" lint --max-diagnostics=none "$path" 2>&1)" || status=$?
  rm -rf "$dir"
  if [[ "$expected" == clean ]]; then
    if [[ $status -eq 0 ]]; then record "$name" yes ''; else record "$name" no "expected clean: $output"; fi
  elif [[ $status -ne 0 && "$output" == *"$expected"* ]]; then
    record "$name" yes ''
  else
    record "$name" no "expected '$expected', exit $status"
  fi
}

suppression_case() {
  local name="$1" file="$2" expected="$3" content="$4" dir actual
  dir="$(new_repo)"
  printf '%s\n' "$content" >"$dir/$file"
  if "$suppressions" "$dir" >/dev/null 2>&1; then actual=0; else actual=1; fi
  rm -rf "$dir"
  if [[ "$actual" == "$expected" ]]; then record "$name" yes ''; else record "$name" no "expected exit $expected, got $actual"; fi
}

css=apps/web/src/case.css

lint_case 'hex color' "$css" 'Raw color.' '.a { color: #fff; }'
lint_case 'hex color as var() fallback' "$css" 'Raw color.' '.a { color: var(--cla-color-text-primary, #000); }'
lint_case 'rgb()' "$css" 'Raw color function.' '.a { background-color: rgb(1 2 3); }'
lint_case 'hsl() inside a gradient' "$css" 'Raw color function.' '.a { background: linear-gradient(hsl(0 0% 0%), var(--cla-color-white)); }'
lint_case 'oklch()' "$css" 'Raw color function.' '.a { outline-color: oklch(70% 0.1 200); }'
lint_case 'named color' "$css" 'Named color.' '.a { color: red; }'
lint_case 'named color in a border shorthand' "$css" 'Named color.' '.a { border: var(--cla-border-width) solid Navy; }'
lint_case 'named color in a custom property' "$css" 'Named color.' '.a { --tone: black; }'
lint_case 'raw padding' "$css" 'Raw length.' '.a { padding: 4px; }'
lint_case 'raw margin in rem' "$css" 'Raw length.' '.a { margin: 0 1rem; }'
lint_case 'raw gap' "$css" 'Raw length.' '.a { gap: 0.5em; }'
lint_case 'raw border radius' "$css" 'Raw length.' '.a { border-top-left-radius: 6px; }'
lint_case 'raw border width' "$css" 'Raw length.' '.a { border: 1px solid var(--cla-color-border-control); }'
lint_case 'raw font size' "$css" 'Raw length.' '.a { font-size: 14px; }'
lint_case 'raw length in calc()' "$css" 'Raw length.' '.a { padding: calc(var(--cla-space-2) + 2px); }'
lint_case 'raw length in a custom property' "$css" 'Raw length.' '.a { --local-gap: 4px; }'
lint_case 'raw shadow' "$css" 'Raw length.' '.a { box-shadow: 0 1px 2px var(--cla-color-navy-950); }'
lint_case 'raw duration in ms' "$css" 'Raw duration.' '.a { transition: opacity 200ms var(--cla-motion-easing-standard); }'
lint_case 'raw duration in s' "$css" 'Raw duration.' '.a { animation-duration: .2s; }'
lint_case 'raw value in the design system outside tokens.css' packages/design-system/src/styles/base.css 'Raw color.' ':root { color: #000; }'

lint_case 'token-only declarations' "$css" clean '.a {
  color: var(--cla-color-text-primary);
  background: var(--cla-gradient-hero);
  padding: var(--cla-space-2) var(--cla-space-4);
  border: var(--cla-border-width) solid var(--cla-color-border-control);
  border-radius: var(--cla-radius-md);
  box-shadow: var(--cla-shadow-sm);
  font-size: var(--cla-text-body-size);
  transition: opacity var(--cla-motion-duration-fast) var(--cla-motion-easing-standard);
  --local: var(--cla-space-2);
}'
lint_case 'keywords and zero' "$css" clean '.a { color: inherit; background: transparent; border-color: currentColor; outline: none; margin: 0; }'
lint_case 'layout values outside the token categories' "$css" clean '@media (min-width: 1024px) { .a { width: 100%; max-width: 72ch; display: grid; } }'
lint_case 'color words outside color properties' "$css" clean '.red { grid-area: navy; animation-name: orange; font-family: system-ui, sans-serif; }'
lint_case 'generated tokens.css is exempt' packages/design-system/src/styles/tokens.css clean ':root { --cla-color-navy-950: #0B1622; --cla-space-1: 4px; --cla-motion-duration-fast: 120ms; }'

suppression_case 'biome-ignore comment in CSS' apps.css 1 '/* biome-ignore lint/plugin: not allowed */
.a { color: #fff; }'
suppression_case 'biome-ignore-all comment in CSS' apps.css 1 '/* biome-ignore-all lint/plugin: not allowed */'
suppression_case 'CSS without suppressions' apps.css 0 '.a { color: var(--cla-color-text-primary); }'
suppression_case 'suppression outside CSS is out of scope' case.ts 0 '// biome-ignore lint/style/useConst: unrelated'

echo "check-css-tokens tests: $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
