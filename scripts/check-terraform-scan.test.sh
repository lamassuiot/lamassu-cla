#!/usr/bin/env bash
# Tests that the Terraform security scan fails closed (A-19, ADR-0012): Checkov must report the
# expected findings in deliberately insecure fixtures, and every inline skip must give a reason.
# Run through `make test-terraform-scan`, which provides CHECKOV_BIN and CHECKOV_FLAGS.
set -euo pipefail

: "${CHECKOV_BIN:?run through make test-terraform-scan}"
: "${CHECKOV_FLAGS:?run through make test-terraform-scan}"

root="$(cd "$(dirname "$0")/.." && pwd)"
fixtures="$root/scripts/testdata/checkov"
pass=0
fail=0

expect_detected() {
  local name="$1" output status=0
  shift
  # shellcheck disable=SC2086 # CHECKOV_FLAGS is a flag list.
  output="$("$CHECKOV_BIN" $CHECKOV_FLAGS --directory "$fixtures/$name" 2>&1)" || status=$?
  if [[ $status -eq 0 ]]; then
    fail=$((fail + 1))
    echo "FAIL: $name: scan exited 0 on an insecure fixture"
    return
  fi
  for id in "$@"; do
    if grep -q "Check: $id:" <<<"$output"; then
      pass=$((pass + 1))
    else
      fail=$((fail + 1))
      echo "FAIL: $name: $id not reported"
    fi
  done
}

expect_detected public-bucket CKV_AWS_53 CKV_AWS_54 CKV_AWS_55 CKV_AWS_56 CKV_AWS_145 CKV_AWS_21
expect_detected unencrypted-key-and-logs CKV_AWS_7 CKV_AWS_158 CKV_AWS_66
expect_detected admin-oidc-role CKV_AWS_62 CKV_AWS_63 CKV_AWS_286

# Exceptions are inline `checkov:skip=<ID>:<reason>` comments; a missing reason fails.
if unjustified="$(grep -rnE 'checkov:skip=[A-Za-z0-9_]+[[:space:]]*(:[[:space:]]*)?$' "$root/infra")"; then
  fail=$((fail + 1))
  echo "FAIL: checkov:skip without a reason:"
  echo "$unjustified"
else
  pass=$((pass + 1))
fi

echo "check-terraform-scan tests: $pass passed, $fail failed"
[[ $fail -eq 0 ]]
