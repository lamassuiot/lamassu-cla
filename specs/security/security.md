# Security Specification

| Field | Value |
| --- | --- |
| Status | Draft — in review; security-owner review required |
| Related | [Threat model](threat-model.md), [data protection](data-protection.md), [architecture](../architecture.md) |

This repository is public. This document describes controls, not environment details. Never add
account IDs, ARNs, resource names, hostnames, or incident details here.

## 1. Principles

- Treat signed agreements, evidence, and contributor data as legal evidence and personal data.
- Deny by default. Every request is authenticated or verified, then authorized, then validated.
- Least privilege for people, functions, and pipelines.
- Fail safe: when coverage or authorization cannot be determined, deny or report pending.
- No secrets in code, configuration files, logs, CI output, or the browser.

## 2. Roles and authorization

Role assignment is open (D-13). Roles are additive; a person can hold several.

| Capability | Anonymous | Contributor | Org representative | Maintainer | Legal admin | Platform admin |
| --- | --- | --- | --- | --- | --- | --- |
| Read published CLA documents | Yes | Yes | Yes | Yes | Yes | Yes |
| Read own status and agreements | — | Yes | Yes | Yes | Yes | Yes |
| Sign an ICLA | — | Yes | Yes | Yes | Yes | Yes |
| Sign a CCLA for an organization | — | — | Yes | — | — | — |
| Manage own organization's contributor list | — | — | Yes | — | — | — |
| Read another contributor's coverage status | — | — | Own organization only | Yes | Yes | Yes |
| Read signed documents and evidence of others | — | — | Own organization's CCLA only | — | Yes | — |
| Approve or reject pending CCLAs | — | — | — | Yes | Yes | — |
| Suspend or revoke agreements and contributors | — | — | — | Suspend only | Yes | — |
| Revalidate contributors or pull requests | — | — | — | Yes | Yes | — |
| Manage project policies | — | — | — | Own repositories | — | Yes |
| Create, publish, deprecate CLA versions | — | — | — | — | Yes | — |
| Configure signing policies | — | — | — | — | Yes | — |
| Configure providers and GitHub App installations | — | — | — | — | — | Yes |
| Read audit log | — | — | — | Own repositories' events | Yes | Yes |

Rules:

- Authorization is enforced in the application layer for every operation. UI guards are not
  security controls.
- Object-level checks: every read or write of a contributor, agreement, or organization verifies
  the caller's relationship to that object. Unauthorized object access returns `404`.
- Platform administrators do not read agreement contents by default. Break-glass access is logged
  and alarmed.
- Role changes take effect on the next request; sessions re-check role membership at a defined
  interval (D-13).

## 3. Authentication

### 3.1 People (GitHub sign-in)

Model choice is open (D-06). Regardless of the choice:

- Authorization-code flow with a cryptographically random `state` bound to a pre-authentication
  cookie, single use, 10-minute expiry, stored hashed in the `ephemeral` table.
- PKCE is used where GitHub supports it for the chosen app type.
- Redirect URIs are registered exactly; no wildcards. Post-login `return_to` accepts only relative
  paths from an allowlist.
- GitHub user access tokens are used only during sign-in to read identity and verified email, then
  discarded. They are never stored or sent to the browser.
- Sessions use an opaque 256-bit random token in a `__Host-` cookie (`HttpOnly`, `Secure`,
  `SameSite=Lax`), stored hashed. Idle and absolute timeouts are configuration (proposed 30 minutes
  idle, 8 hours absolute). Sessions rotate on sign-in and role change and are revoked on sign-out.
- CSRF: `SameSite=Lax` plus a session-bound `X-CSRF-Token` header and `Origin` check on state-changing
  requests.

### 3.2 GitHub App

- App JWTs are signed with the private key stored in Secrets Manager and are valid for at most 10
  minutes. Installation tokens are requested per job with the minimum permissions and repositories
  needed, cached in memory only until expiry, and never logged.
- Webhooks are verified with HMAC-SHA256 over the raw body using the webhook secret, compared in
  constant time, before parsing.
- Requested App permissions (proposed): Checks read and write, Pull requests read, Contents read
  (commits only), Metadata read, Members read (team-based roles, if chosen in D-13).

### 3.3 Signing provider

- Provider credentials (for DocuSign: integration key and private key for JWT grant) live in Secrets
  Manager.
- Callbacks are verified with the provider's mechanism (for DocuSign Connect: HMAC). The callback is
  then treated only as a hint; status and documents are re-read from the provider API.

### 3.4 Deployments

- GitHub Actions assume AWS roles with OIDC. Trust policies restrict `sub` to this repository and,
  for deployments, to the specific GitHub environment.
- Production requires a GitHub environment with required reviewers.
- No long-lived AWS access keys exist for CI.

## 4. Data protection

- Encryption at rest with customer-managed KMS keys, separated by data class: `core` data, `audit`
  data, evidence objects, secrets.
- TLS 1.2 or later for all traffic. API Gateway and CloudFront use AWS-managed security policies.
- S3 Block Public Access on every bucket and at account level. Bucket policies deny non-TLS
  requests and unencrypted uploads.
- Versioning on all buckets; Object Lock on `evidence` ([ADR-0005](../../docs/decisions/0005-configurable-object-lock-retention.md)).
- Raw email addresses are never persisted or logged. Persisted lookup values are keyed HMACs
  computed with an AWS KMS HMAC key (D-19, technical default pending DPO confirmation).
- Details: [data protection](data-protection.md).

## 5. Input validation

- Requests are validated against the OpenAPI schema and then by domain rules.
- Unknown fields are rejected. Sizes, lengths, and formats are bounded.
- Webhook bodies have a maximum size and are parsed only after signature verification.
- Uploaded files (if any are introduced) are type-checked and size-limited; documents from
  providers are checked for expected content type and hashed on receipt.
- Output encoding is handled by React; `dangerouslySetInnerHTML` is prohibited.
- The portal sets a strict Content Security Policy, `X-Content-Type-Options`, `Referrer-Policy`, and
  frame-ancestors restrictions.

## 6. Idempotency and replay

- Webhook deliveries and provider callbacks are deduplicated in the `ephemeral` table.
- State transitions are conditional on current state and version, so replays cannot regress state.
- Mutating API operations use `Idempotency-Key` or `If-Match` ([API conventions](../api/conventions.md#6-idempotency)).

## 7. Logging and audit

- Logs are structured and contain request IDs, actor IDs, and outcomes only. They never contain
  tokens, cookies, secrets, emails, names, document contents, or full webhook payloads.
- Audit events are append-only ([domain model](../domain-model.md#212-audit-event)) and exported to
  locked storage.
- CloudTrail records management events and data events on the `evidence` bucket.
- Security-relevant events (failed signature verification, authorization denials, break-glass
  access) produce metrics and alarms.

## 8. Rate limiting and abuse

- API Gateway throttling per stage and route; stricter limits on sign-in and signing-session
  creation.
- Optional AWS WAF with rate-based rules and managed rule groups (D-16).
- Signing sessions per contributor are limited to prevent envelope-creation abuse and cost.

## 9. Infrastructure security

- One IAM role per Lambda function with only the actions and resources it needs.
- The `audit` table denies `UpdateItem` and `DeleteItem` to all application roles.
- Object Lock bypass permission is not granted to application or deployment roles.
- Terraform is scanned in CI with Trivy (A-19) and TFLint.

## 10. Supply chain

- GitHub Actions pinned by commit SHA (existing practice).
- Dependabot for GitHub Actions, Go modules, npm, and Terraform providers, added as manifests appear.
- `govulncheck`, dependency review, CodeQL, gitleaks, and OpenSSF Scorecard in CI.
- Lockfiles committed; installs in CI use frozen lockfiles.

## 11. Pull-request security checklist

Every PR that touches code or infrastructure confirms:

- [ ] No secrets, personal data, or real agreements in code, fixtures, or logs.
- [ ] New operations declare authentication, roles, and error responses in OpenAPI.
- [ ] Object-level authorization is tested for new data access.
- [ ] Inputs are validated and bounded.
- [ ] New IAM permissions are least privilege and justified.
- [ ] Webhook and callback handling is verified and idempotent.
- [ ] Threat model updated if a trust boundary or asset changed.
