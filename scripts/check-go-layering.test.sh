#!/usr/bin/env bash
# Verifies that golangci-lint enforces the hexagonal layering (F-01 AC-01-4, F01-T2).
# Each case adds a violating file to a temporary copy of services/api and expects a depguard failure.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
lint="${GOLANGCI_LINT:-$root/.tools/bin/golangci-lint}"
module="github.com/lamassuiot/lamassu-cla/services/api"
failures=0

# Arguments: name, expected result (pass or fail), package directory, import path.
run_case() {
  local name="$1" expected="$2" dir="$3" import="$4"
  local work status=0 pkg
  work="$(mktemp -d)"
  cp -R "$root/services/api/." "$work/"
  pkg="$(basename "$dir")"
  printf 'package %s\n\nimport _ "%s"\n' "$pkg" "$import" > "$work/$dir/layering_probe.go"
  (cd "$work" && "$lint" run --enable-only depguard ./... > "$work/lint.log" 2>&1) || status=$?
  local result="pass"
  if [[ "$status" -ne 0 ]]; then
    result="error"
    grep -q '(depguard)' "$work/lint.log" && result="fail"
  fi
  if [[ "$result" == "$expected" ]]; then
    echo "ok: $name"
  else
    echo "FAIL: $name (expected $expected, got $result)"
    sed 's/^/    /' "$work/lint.log"
    failures=$((failures + 1))
  fi
  rm -rf "$work"
}

run_case "domain may not import application" fail internal/domain "$module/internal/application"
run_case "domain may not import adapters" fail internal/domain "$module/internal/adapters/aws"
run_case "domain may not perform network I/O" fail internal/domain "net/http"
run_case "domain may not access the OS" fail internal/domain "os"
run_case "ports may not import application" fail internal/ports "$module/internal/application"
run_case "application may not import adapters" fail internal/application "$module/internal/adapters/signing"
run_case "application may not import handlers" fail internal/application "$module/internal/handlers"
run_case "handlers may not import adapters" fail internal/handlers "$module/internal/adapters/github"
run_case "middleware may not import adapters" fail internal/middleware "$module/internal/adapters/aws"
run_case "domain may import standard library" pass internal/domain "strings"
run_case "application may import ports" pass internal/application "$module/internal/ports"
run_case "handlers may import application" pass internal/handlers "$module/internal/application"
run_case "adapters may import ports" pass internal/adapters/aws "$module/internal/ports"

if ((failures)); then
  echo "${failures} layering test(s) failed."
  exit 1
fi
echo "All layering tests passed."
