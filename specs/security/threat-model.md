# Threat Model

| Field | Value |
| --- | --- |
| Status | Draft — in review; security-owner review required |
| Method | STRIDE per trust boundary |
| Related | [Security specification](security.md), [architecture](../architecture.md), [test scenarios](../test-scenarios/README.md) |

## 1. Assets

| Asset | Why it matters |
| --- | --- |
| Signed agreements and evidence | Legal evidence of contribution rights. |
| Agreement and coverage state | Determines whether pull requests pass. |
| Contributor and signer personal data | Privacy obligations. |
| Audit events | Proof of who did what and when. |
| GitHub App private key and webhook secret | Control over repositories where the App is installed. |
| Signing-provider credentials | Ability to create envelopes and read signed documents. |
| Published CLA text and hashes | Defines what contributors agreed to. |
| Terraform state | Infrastructure configuration and resource metadata; unauthorized access or modification can expose environment details or alter deployments. |

## 2. Trust boundaries

1. Internet to API Gateway (portal users, anonymous visitors).
2. GitHub to webhook endpoint.
3. Signing provider to callback endpoint.
4. Lambda functions to AWS services.
5. Lambda functions to GitHub and provider APIs.
6. GitHub Actions to AWS (deployment).
7. GitHub Actions and operators to the Terraform state backend in the Infrastructure/Tooling
   account (target state, D-24).

## 3. Threats and mitigations

STRIDE: **S**poofing, **T**ampering, **R**epudiation, **I**nformation disclosure, **D**enial of
service, **E**levation of privilege.

| ID | Threat | STRIDE | Mitigations | Residual risk | Tests |
| --- | --- | --- | --- | --- | --- |
| T-01 | Forged GitHub webhook | S, T | HMAC-SHA256 verification on raw body before parsing; constant-time compare; unknown installations ignored. | Webhook secret compromise; mitigated by rotation procedure. | TS-11 |
| T-02 | Forged provider callback | S, T | Provider verification (HMAC); callback treated as a hint; status and documents re-read from provider API with platform credentials. | Provider account compromise. | TS-04 |
| T-03 | Replay of webhooks or callbacks | T | Delivery deduplication; conditional state transitions; stale events cannot regress state. | None significant. | TS-03 |
| T-04 | Unauthorized access to agreements | I | Object-level authorization; `404` for others' objects; documents served only through authorized API; no public bucket access. | Authorization bugs; mitigated by tests per operation. | TS-12 |
| T-05 | Unauthorized CCLA administration | E, S | Representative authority reviewed by maintainer or legal admin; contributor lists changed only by the representative or authorized roles; every change audited. | Social engineering of reviewers. | TS-10, TS-13 |
| T-06 | GitHub App private-key compromise | S, E | Key in Secrets Manager with restricted IAM; never logged; minimal App permissions; CloudTrail on secret access; documented rotation and revocation. | Window between compromise and rotation. | Operations runbook |
| T-07 | Replacement or deletion of evidence | T, R | Object Lock; versioning; hashes recorded in DynamoDB and audit; no delete or bypass permission for application roles; CloudTrail data events. | Governance-mode bypass by privileged admins outside production. | Infrastructure tests |
| T-08 | Privilege escalation in the portal | E | Server-side role checks; roles from trusted source (D-13); session rotation on role change; admin actions audited. | Misconfigured role mapping. | Authorization test suite |
| T-09 | Organization impersonation | S | No coverage from email domains ([ADR-0006](../../docs/decisions/0006-explicit-ccla-authorization.md)); reviewer approval of representative authority; legal name and jurisdiction recorded. | Convincing forged authority. Legal process applies. | TS-09 domain-mismatch case |
| T-10 | Accidental exposure of signer data | I | Data minimization; encrypted signer fields; no personal data in logs, keys, metrics, or Check Run output; public catalogue exposes no signer data. | Human error in new code; mitigated by review checklist. | Log-redaction tests |
| T-11 | Malicious pull-request author | S, D | Coverage based on GitHub user IDs resolved by the API, not on commit email alone; unresolved identities produce `ACTION_REQUIRED`; Check Run output contains no private data; event floods are rate-limited and deduplicated. | Commit author spoofing with an email of a covered person; mitigated by D-14 rules. | TS-07, TS-08 |
| T-12 | Provider outage | D | Clear retryable errors; queued processing with retries; reconciliation job; no activation without confirmation. | Delayed activations. | TS-15 |
| T-13 | Duplicate callbacks | T | Deduplication and idempotent transitions. | None significant. | TS-03 |
| T-14 | Partial workflow failure | T, R | S3 first, then DynamoDB transaction; reconciliation detects orphans; agreements stay `VERIFYING` until complete; retries are idempotent. | Manual follow-up for persistent failures. | TS-05, TS-06 |
| T-15 | Session theft or CSRF | S, E | `__Host-` HttpOnly Secure cookies; CSRF token and Origin checks; strict CSP; short session lifetimes. | Compromised user device. | Security tests |
| T-16 | Open redirect in sign-in or signing return | S | Exact redirect URI registration; relative-path allowlist for `return_to`; one-time return nonce bound to the signing session. | None significant. | Security tests |
| T-17 | Compromised CI or deployment pipeline | E, T | OIDC trust restricted to repository and environment; production approval; SHA-pinned actions; protected branches with signed commits and Code Owner review. | Maintainer account compromise. | Configuration review |
| T-18 | Disclosure through the public repository | I | Policy in `AGENTS.md`; gitleaks; private-URL scan; synthetic fixtures; no environment identifiers committed. | Human error. | CI |
| T-19 | Disclosure or tampering of Terraform state *(proposed; target state, D-24)* | I, T, E | State is stored in a dedicated Infrastructure/Tooling account: never in application workload accounts, and never in the Control Tower management, Log Archive, or Audit accounts. Separate state buckets and separate customer-managed KMS keys for `dev`, `staging`, and `production`, each with Block Public Access, SSE-KMS, versioning, and a TLS-only policy. Unique state keys per repository, environment, and stack (`lamassu-cla/<environment>/<stack>/terraform.tfstate`), with S3-native locking through the corresponding `.tflock` object. Separate roles for backend state access and for workload-account deployment, assumed through GitHub OIDC; each is restricted to its environment and state prefix. The current `infra/bootstrap` keeps state in each workload account (C-01 in `specs/STATE.md`). | A compromise of the tooling account, or of an identity with administrative access to it, could expose or alter the state of several environments. State may contain sensitive values if future resources record them. | F-03 mocked-provider tests; Checkov (ADR-0012); IAM policy review |

## 4. Review cadence

Update this model when a trust boundary, asset, or authentication mechanism changes, and before
production launch. Security-owner approval is required for changes.
