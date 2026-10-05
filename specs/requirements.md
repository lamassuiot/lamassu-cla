# Product Requirements

| Field | Value |
| --- | --- |
| Status | Draft — in review |
| Owner | Repository owner |
| Related | [Architecture](architecture.md), [domain model](domain-model.md), [decision log](../docs/decisions/decision-log.md) |

Requirement IDs are stable. Feature specifications and tests reference them. Numeric targets marked
**(proposed)** need confirmation from the repository owner.

## 1. Purpose

Lamassu open-source projects require a Contributor License Agreement (CLA) before accepting
contributions. Verification is currently manual. The CLA Management Platform provides a secure and
auditable workflow to:

- Publish CLA versions.
- Collect and store signed Individual (ICLA) and Corporate (CCLA) agreements.
- Manage contributor and organization coverage.
- Validate GitHub pull requests through a GitHub App.

## 2. Scope

### In scope

- Public catalogue of CLA documents and versions.
- GitHub sign-in for people.
- ICLA and CCLA signing through external electronic-signature providers, starting with DocuSign.
- Secure storage of signed documents and evidence.
- Organization coverage with explicit authorized-contributor lists.
- Pull-request validation with GitHub Check Runs.
- Maintainer, legal-administrator, and platform-administrator tooling.
- Audit trail for every state change.

### Out of scope

- Drafting or interpreting agreement text (Legal supplies it).
- Implementing electronic-signature cryptography.
- Repositories outside GitHub.
- Payments or commercial licensing.
- Automatic CCLA coverage based on email domains ([ADR-0006](../docs/decisions/0006-explicit-ccla-authorization.md)).

## 3. Users and roles

| Role | Description |
| --- | --- |
| Anonymous visitor | Reads public CLA documents and signing instructions. |
| Contributor | Signs in with GitHub, signs an ICLA, sees their own status and coverage. |
| Organization representative | Signs a CCLA on behalf of an organization and manages its authorized contributor list. |
| Maintainer | Reviews pending CCLAs, checks contributor coverage, triggers revalidation, manages project policies for repositories they maintain. |
| Legal administrator | Manages CLA documents and versions, signing policies, legal holds, and agreement revocation. Reads private agreements and evidence. |
| Platform administrator | Manages provider configuration, GitHub App installations, and operational settings. Does not read agreement contents by default. |
| GitHub App (system) | Sends webhooks and receives Check Run updates. |
| Signing provider (system) | Sends signing callbacks and serves signed documents. |

Role assignment is open (D-13). The authorization matrix is in the
[security specification](security/security.md).

## 4. Functional requirements

### 4.1 Public CLA catalogue (FR-CAT)

| ID | Requirement |
| --- | --- |
| FR-CAT-01 | Anyone can list the currently published ICLA and CCLA documents without signing in. |
| FR-CAT-02 | Each document has a stable identifier, a type (ICLA or CCLA), a version, a language, an effective date, and a status: Draft, Published, Deprecated, or Archived. |
| FR-CAT-03 | Every published document has a SHA-256 hash, shown publicly. |
| FR-CAT-04 | A published version is immutable. Corrections create a new version. |
| FR-CAT-05 | Version history is publicly visible for Published, Deprecated, and Archived versions. Drafts are visible only to legal administrators. |
| FR-CAT-06 | Only legal administrators can create drafts and publish, deprecate, or archive versions. |
| FR-CAT-07 | Public responses never contain signer or contributor information. |
| FR-CAT-08 | The published text matches the canonical text in the public repository, verified by hash ([ADR-0002](../docs/decisions/0002-public-cla-text-and-private-evidence.md)). |

### 4.2 Contributors (FR-CON)

| ID | Requirement |
| --- | --- |
| FR-CON-01 | A contributor is identified by their immutable GitHub user ID. The GitHub login is stored for display and refreshed on sign-in. |
| FR-CON-02 | Contributor status is Pending, Active, Suspended, or Revoked. |
| FR-CON-03 | A contributor can see their own agreements, coverage, and status. They cannot see other contributors' data. |
| FR-CON-04 | Raw email addresses are never persisted or logged. Email lookups use keyed HMAC values (D-19). |
| FR-CON-05 | Every change to a contributor's status is audited with the actor and reason. |
| FR-CON-06 | Maintainers and administrators can search contributors by GitHub login or user ID. Results show only the data their role allows. |

### 4.3 ICLA workflow (FR-ICLA)

| ID | Requirement |
| --- | --- |
| FR-ICLA-01 | The contributor signs in with GitHub before starting an ICLA. |
| FR-ICLA-02 | The contributor reviews the current published ICLA version, identified by version and hash, before signing. |
| FR-ICLA-03 | The platform creates a signing envelope through the configured provider and redirects the contributor to it. |
| FR-ICLA-04 | Provider callbacks are verified before processing. The platform re-reads envelope status from the provider instead of trusting callback contents. |
| FR-ICLA-05 | After completion, the platform downloads the signed document and evidence package, stores them in private storage, and records their SHA-256 hashes. |
| FR-ICLA-06 | The agreement becomes Active only after every validation step succeeds, including the signature-level check ([ADR-0004](../docs/decisions/0004-provider-neutral-signing-and-signature-level.md)). |
| FR-ICLA-07 | The active ICLA is linked to the contributor's GitHub user ID. |
| FR-ICLA-08 | Abandoned, declined, expired, or failed signing sessions leave no active agreement and can be restarted. |
| FR-ICLA-09 | Duplicate callbacks have no additional effect. |
| FR-ICLA-10 | Each step produces an audit event. |

### 4.4 CCLA workflow (FR-CCLA)

| ID | Requirement |
| --- | --- |
| FR-CCLA-01 | A CCLA records the organization's legal name, jurisdiction, authorized representative, the representative's identity, effective date, and agreement version. |
| FR-CCLA-02 | The representative states their authority during signing. A maintainer or legal administrator reviews it before the CCLA becomes Active. |
| FR-CCLA-03 | A CCLA may require multiple signers, as configured by the signing policy. |
| FR-CCLA-04 | The organization maintains an authorized-contributor list identified by GitHub user ID. |
| FR-CCLA-05 | Coverage may be limited to specific repositories or projects. |
| FR-CCLA-06 | Coverage is never inferred from an email domain. |
| FR-CCLA-07 | Authorized maintainers can suspend or revoke a CCLA or remove individual contributors. Coverage ends immediately for future validations. |
| FR-CCLA-08 | Authorized contributor lists are revalidated periodically. The interval is a configuration value. |
| FR-CCLA-09 | A contribution covered by a CCLA is not also accepted under an ICLA for the same contributor and repository. Precedence rules are defined in the Phase 8 validation algorithm. |

### 4.5 GitHub pull-request validation (FR-GH)

| ID | Requirement |
| --- | --- |
| FR-GH-01 | The GitHub App receives pull-request, push, installation, and installation-repository events. |
| FR-GH-02 | Webhook signatures are verified before any processing. Unsigned or invalid requests are rejected and counted. |
| FR-GH-03 | Duplicate deliveries are detected by delivery ID and have no additional effect. |
| FR-GH-04 | The validation identifies every relevant contributor in the pull request: the PR author and commit authors, and optionally co-authors from commit trailers (D-14). |
| FR-GH-05 | Each contributor passes when they have an active ICLA or active CCLA coverage for the repository. |
| FR-GH-06 | The App creates or updates a Check Run with outcome passed, pending (action required), or failed, and a link to the signing portal for uncovered contributors. |
| FR-GH-07 | Repositories without CLA enforcement enabled are never blocked. |
| FR-GH-08 | Maintainers can request revalidation of a pull request. |
| FR-GH-09 | The App uses installation access tokens only; no personal access tokens are stored. |
| FR-GH-10 | The validation algorithm is specified and approved before implementation (Phase 8). |

### 4.6 Signing providers (FR-SIGN)

| ID | Requirement |
| --- | --- |
| FR-SIGN-01 | Application logic depends only on a provider-neutral signing port. |
| FR-SIGN-02 | The first adapter targets DocuSign: envelope creation with one or more signers, return URLs, status retrieval, callbacks, signed PDF download, and certificate or audit-trail download. |
| FR-SIGN-03 | Provider errors are mapped to provider-neutral error categories. |
| FR-SIGN-04 | The required signature level and evidence are configured per agreement type. The platform never claims a qualified signature unless the configured service is qualified. |
| FR-SIGN-05 | Provider credentials are stored in AWS Secrets Manager and never in code, configuration files, or logs. |

### 4.7 Project policies (FR-POL)

| ID | Requirement |
| --- | --- |
| FR-POL-01 | Each repository has a policy: CLA required or not, accepted agreement types, covered organizations, and status. |
| FR-POL-02 | Repositories without a policy are treated as not enforced. |
| FR-POL-03 | Policy changes are audited and take effect for subsequent validations. |

### 4.8 Audit (FR-AUD)

| ID | Requirement |
| --- | --- |
| FR-AUD-01 | Every state change records an audit event: event type, actor type and ID, entity, timestamp, request ID, and value hashes. |
| FR-AUD-02 | Audit events are append-only. No role can modify or delete them through the application. |
| FR-AUD-03 | Authorized roles can view and filter audit events. |
| FR-AUD-04 | Audit events can be exported to protected storage for long-term evidence. |

### 4.9 Portal (FR-UI)

| ID | Requirement |
| --- | --- |
| FR-UI-01 | Screens: public landing page, GitHub sign-in, contributor dashboard, ICLA flow, CCLA flow, organization administration, agreement status, repository coverage, maintainer dashboard, pending-review queue, search, audit log, project policies, GitHub App configuration, and provider configuration (administrators only). |
| FR-UI-02 | Every data-driven view handles loading, empty, error, and permission-denied states. |
| FR-UI-03 | Signing flows handle expired sessions, pending provider callbacks, and revalidation requirements. |
| FR-UI-04 | Light and dark themes built from shared design tokens ([design tokens](ux/design-tokens.md)). |
| FR-UI-05 | Responsive layout from small mobile screens to wide desktop screens. |

## 5. Non-functional requirements

| ID | Category | Requirement |
| --- | --- | --- |
| NFR-01 | Security | Controls in the [security specification](security/security.md) are mandatory. |
| NFR-02 | Privacy | Personal data is minimized, encrypted, and retained per the [data-protection notes](security/data-protection.md). |
| NFR-03 | Integrity | Signed documents and evidence are immutable for their retention period ([ADR-0005](../docs/decisions/0005-configurable-object-lock-retention.md)). |
| NFR-04 | Accessibility | The portal meets WCAG 2.2 level AA. |
| NFR-05 | Availability | **(proposed)** 99.5% monthly availability for the public catalogue and GitHub validation in production. |
| NFR-06 | Latency | **(proposed)** Webhook endpoints acknowledge within 2 seconds at p99; Check Runs are updated within 60 seconds of a PR event at p95. |
| NFR-07 | Recovery | **(proposed)** Recovery point objective 1 hour (DynamoDB point-in-time recovery) and recovery time objective 4 hours. |
| NFR-08 | Observability | Structured logs with request correlation IDs, metrics, and alarms for every external integration. |
| NFR-09 | Idempotency | Every webhook, callback, and mutating API operation is idempotent. |
| NFR-10 | Portability | Replacing the signing provider requires no domain-layer changes. |
| NFR-11 | Maintainability | Infrastructure is managed in Terraform, and the API contract in OpenAPI. |
| NFR-12 | Fail-safe | If the platform cannot determine coverage, the Check Run is pending, never passed. |

## 6. Constraints

- Public repository, AGPLv3 license, copyright LKS S. Coop.
- AWS: API Gateway HTTP APIs, Lambda (Go), DynamoDB, S3, KMS, Secrets Manager, CloudWatch.
- Frontend: Vite, React, TypeScript, React Router, TanStack Query; Node.js 24.
- Terraform for all infrastructure; GitHub Actions with OIDC for deployments.
- Repository governance in [`AGENTS.md`](../AGENTS.md) and [`CONTRIBUTING.md`](../CONTRIBUTING.md).

## 7. Success criteria

- Contributors can sign an ICLA end to end without maintainer involvement.
- Pull requests in enforced repositories receive a correct Check Run without manual checks.
- Maintainers no longer verify CLA coverage manually for enforced repositories.
- Every active agreement can be traced to its signed document, evidence, hash, and audit events.

## 8. Assumptions

- The signing provider offers webhooks or a status API sufficient to confirm completion.
- GitHub Check Runs can be required by repository rulesets in enforced repositories.
- Legal supplies agreement text and retention decisions before the corresponding phases.

## 9. Required test scenarios

The minimum scenarios are listed in the [test-scenario catalogue](test-scenarios/README.md).
