#!/usr/bin/env bash
# Applies baseline GitHub settings to a repository generated from this template.
# Template-only: delete this file after instantiation. See TEMPLATE.md.
set -euo pipefail

usage() {
  echo "Usage: $0 OWNER/REPO [--yes]" >&2
  exit 2
}

[[ $# -ge 1 && $# -le 2 ]] || usage
repo="$1"
[[ "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || usage
[[ $# -eq 1 || "$2" == "--yes" ]] || usage
owner="${repo%%/*}"

ruleset_name="Protect default branch"
# Job names from the template workflows; update them if the workflows change.
required_checks=(
  "Validate PR title format"
  "Validate repository template"
  "GitHub Actions security lint"
  "Secret scan"
  "Dependency review"
)
required_labels=(bug enhancement needs-triage)
required_issue_types=(Bug Feature Task)

failures=()

step() {
  local description="$1"
  shift
  echo "==> ${description}"
  if ! "$@"; then
    echo "    FAILED: ${description}" >&2
    failures+=("$description")
  fi
}

configure_merging() {
  gh api --silent -X PATCH "repos/${repo}" \
    -F allow_squash_merge=true \
    -F allow_merge_commit=false \
    -F allow_rebase_merge=false \
    -f squash_merge_commit_title=PR_TITLE \
    -f squash_merge_commit_message=COMMIT_MESSAGES \
    -F delete_branch_on_merge=true
}

enable_private_vulnerability_reporting() {
  gh api --silent -X PUT "repos/${repo}/private-vulnerability-reporting"
}

enable_dependabot_security() {
  gh api --silent -X PUT "repos/${repo}/vulnerability-alerts" &&
    gh api --silent -X PUT "repos/${repo}/automated-security-fixes"
}

enable_secret_scanning() {
  gh api --silent -X PATCH "repos/${repo}" \
    -f 'security_and_analysis[secret_scanning][status]=enabled' \
    -f 'security_and_analysis[secret_scanning_push_protection][status]=enabled'
}

apply_ruleset() {
  local checks_json="" check existing_id ruleset_json
  for check in "${required_checks[@]}"; do
    checks_json+="${checks_json:+,}{\"context\":\"${check}\"}"
  done

  ruleset_json=$(
    cat <<EOF
{
  "name": "${ruleset_name}",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["~DEFAULT_BRANCH"], "exclude": [] } },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "required_signatures" },
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 1,
        "require_code_owner_review": true,
        "dismiss_stale_reviews_on_push": true,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false,
        "allowed_merge_methods": ["squash"]
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "strict_required_status_checks_policy": false,
        "required_status_checks": [${checks_json}]
      }
    }
  ]
}
EOF
  )

  existing_id=$(gh api "repos/${repo}/rulesets" --jq ".[] | select(.name == \"${ruleset_name}\") | .id")
  if [[ -n "$existing_id" ]]; then
    gh api --silent -X PUT "repos/${repo}/rulesets/${existing_id}" --input - <<<"$ruleset_json"
  else
    gh api --silent -X POST "repos/${repo}/rulesets" --input - <<<"$ruleset_json"
  fi
}

# Labels and Issue Types are organization-level; only verify them, never create them here.
verify_labels() {
  local existing label missing=0
  existing=$(gh label list -R "$repo" --limit 500 --json name --jq '.[].name')
  for label in "${required_labels[@]}"; do
    if ! grep -qxF "$label" <<<"$existing"; then
      echo "    Missing label: ${label}" >&2
      missing=1
    fi
  done
  return "$missing"
}

verify_issue_types() {
  local existing issue_type missing=0
  existing=$(gh api "orgs/${owner}/issue-types" --jq '.[].name')
  for issue_type in "${required_issue_types[@]}"; do
    if ! grep -qxF "$issue_type" <<<"$existing"; then
      echo "    Missing organization Issue Type: ${issue_type}" >&2
      missing=1
    fi
  done
  return "$missing"
}

gh auth status >/dev/null

cat <<EOF
This changes settings of https://github.com/${repo}:
  - squash merging only (PR title as commit title); delete head branches after merge
  - private vulnerability reporting; Dependabot alerts and security updates
  - secret scanning and push protection
  - ruleset "${ruleset_name}" on the default branch
It also verifies the required labels and organization Issue Types.
EOF

if [[ "${2:-}" != "--yes" ]]; then
  read -r -p "Continue? [y/N] " answer
  [[ "$answer" =~ ^[Yy]$ ]] || {
    echo "Aborted."
    exit 1
  }
fi

step "Merge settings" configure_merging
step "Private vulnerability reporting" enable_private_vulnerability_reporting
step "Dependabot alerts and security updates" enable_dependabot_security
step "Secret scanning and push protection" enable_secret_scanning
step "Default branch ruleset" apply_ruleset
step "Required labels" verify_labels
step "Organization Issue Types" verify_issue_types

if [[ ${#failures[@]} -gt 0 ]]; then
  echo
  echo "Completed with failures:" >&2
  printf '  - %s\n' "${failures[@]}" >&2
  echo "Rulesets and private-repository security features may require a public repository or a paid plan." >&2
  exit 1
fi

echo
echo "Done. Complete the remaining manual steps in TEMPLATE.md."
