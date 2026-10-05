#!/usr/bin/env bash
# Tests for check-private-urls.sh. Private host labels are assembled at runtime so this file passes the scan.
set -euo pipefail

script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-private-urls.sh"
internal="inter""nal"
intranet="intra""net"
corp="co""rp"
failures=0

# Arguments: name, expected exit code, sample file content, optional allowlist content.
run_case() {
  local name="$1" expected="$2" content="$3" allowlist="${4-}"
  local dir status=0
  dir="$(mktemp -d)"
  git -C "$dir" init -q
  printf '%s\n' "$content" > "$dir/sample.md"
  if [[ -n "$allowlist" ]]; then
    mkdir -p "$dir/.github"
    printf '%s\n' "$allowlist" > "$dir/.github/private-url-allowlist.txt"
  fi
  "$script" "$dir" > /dev/null 2>&1 || status=$?
  rm -rf "$dir"
  if [[ "$status" -eq "$expected" ]]; then
    echo "ok: $name"
  else
    echo "FAIL: $name (expected exit $expected, got $status)"
    failures=$((failures + 1))
  fi
}

run_case "blocks private host in URL" 1 "See https://jira.${internal}/browse/X-1."
run_case "blocks private label as first host label" 1 "https://${internal}.example.com/path"
run_case "blocks corp host" 1 "http://wiki.${corp}/page"
run_case "blocks intranet host" 1 "https://${intranet}.example.org"
run_case "blocks private host with userinfo and port" 1 "ssh://git@git.${internal}:2222/repo.git"
run_case "allows Go internal package import path" 0 "import \"github.com/lamassuiot/lamassu-cla/services/api/${internal}/domain\""
run_case "allows internal path segment in public URL" 0 "https://pkg.go.dev/example.com/mod/${internal}/x"
run_case "allows dotted internal identifier" 0 "Use ${internal}.Config from the package."
run_case "allows localhost development URL" 0 "Open http://localhost:3000 in a browser."
run_case "allows loopback development URL" 0 "http://127.0.0.1:8080/v1/health"
run_case "allows similar public hostname" 0 "https://${internal}ize.example.com"
run_case "allows similar public label" 0 "https://my-${corp}.example.com"
run_case "allows justified allowlist entry" 0 "https://docs.${internal}.example.org" \
  "# Public documentation site with a private-looking label.
https://docs.${internal}.example.org"
run_case "rejects allowlist entry without justification" 2 "https://docs.${internal}.example.org" \
  "https://docs.${internal}.example.org"

if ((failures)); then
  echo "${failures} test(s) failed."
  exit 1
fi
echo "All private-URL scan tests passed."
