#!/usr/bin/env bash
# Tests for scripts/check-ts-directives.sh.
set -euo pipefail

script="$(cd "$(dirname "$0")" && pwd)/check-ts-directives.sh"
pass=0
fail=0

run_case() {
  local name="$1" expected="$2" content="$3" dir
  dir="$(mktemp -d)"
  git -C "$dir" init -q
  printf '%s\n' "$content" >"$dir/case.ts"
  if "$script" "$dir" >/dev/null 2>&1; then actual=0; else actual=1; fi
  rm -rf "$dir"
  if [[ "$actual" == "$expected" ]]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "FAIL: $name (expected exit $expected, got $actual)"
  fi
}

run_case 'triple-slash path reference' 1 '/// <reference path="./other.ts" />'
run_case 'triple-slash types reference' 1 '  /// <reference types="vite/client" />'
run_case 'ts-nocheck' 1 '// @ts-nocheck'
run_case 'bare ts-expect-error' 1 $'// @ts-expect-error\nconst a: number = "x";'
run_case 'bare ts-expect-error with colon' 1 $'// @ts-expect-error:\nconst a: number = "x";'
run_case 'block ts-expect-error without text' 1 $'/* @ts-expect-error */\nconst a: number = "x";'
run_case 'described ts-expect-error' 0 $'// @ts-expect-error: fixture checks the invalid assignment\nconst a: number = "x";'
run_case 'plain comment mentioning reference' 0 '// see <reference> docs'
run_case 'ordinary triple-slash comment' 0 '/// A doc comment.'
run_case 'clean file' 0 'export const a = 1;'

echo "check-ts-directives tests: $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
