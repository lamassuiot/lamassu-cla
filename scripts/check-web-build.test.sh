#!/usr/bin/env bash
# Tests for scripts/check-web-build.sh.
set -euo pipefail

script="$(cd "$(dirname "$0")" && pwd)/check-web-build.sh"
pass=0
fail=0

# run_case NAME EXPECTED_EXIT FILE CONTENT
run_case() {
  local name="$1" expected="$2" file="$3" content="$4" dir actual
  dir="$(mktemp -d)"
  mkdir -p "$dir/assets"
  printf '<!doctype html><script type="module" src="/assets/index.js"></script>\n' >"$dir/index.html"
  printf '%s\n' "$content" >"$dir/$file"
  if "$script" "$dir" >/dev/null 2>&1; then actual=0; else actual=1; fi
  rm -rf "$dir"
  if [[ "$actual" == "$expected" ]]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "FAIL: $name (expected exit $expected, got $actual)"
  fi
}

run_case 'woff2 font file' 1 assets/inter.woff2 'binary'
run_case 'ttf font file' 1 assets/mono.TTF 'binary'
run_case 'remote stylesheet link' 1 index.html '<link rel="stylesheet" href="https://fonts.example.com/css">'
run_case 'protocol-relative script' 1 index.html '<script src="//cdn.example.com/a.js"></script>'
run_case 'remote url() in CSS' 1 assets/index.css '.a{background:url(https://cdn.example.com/a.png)}'
run_case 'quoted protocol-relative url() in CSS' 1 assets/index.css '.a{background:url("//cdn.example.com/a.png")}'
run_case 'remote @import in CSS' 1 assets/index.css '@import "https://fonts.example.com/css";'
run_case '@font-face in CSS' 1 assets/index.css '@font-face{font-family:x;src:url(/x.woff2)}'
run_case 'remote dynamic import' 1 assets/index.js 'import("https://cdn.example.com/m.js")'
run_case 'local assets only' 0 assets/index.css '.a{background:url(/assets/a.svg);color:var(--cla-color-text-primary)}'
run_case 'URL strings that are not load points' 0 assets/index.js 'const ns="http://www.w3.org/2000/svg";const doc="https://react.dev/errors/";'
run_case 'local dynamic import' 0 assets/index.js 'import("./chunk.js")'

missing="$(mktemp -d)/dist"
if "$script" "$missing" >/dev/null 2>&1; then
  fail=$((fail + 1))
  echo 'FAIL: missing build output (expected exit 1, got 0)'
else
  pass=$((pass + 1))
fi

echo "check-web-build tests: $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
