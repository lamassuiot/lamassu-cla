# Decision Log

This log records every open and settled decision for the CLA Management Platform. Architecture
decision records (ADRs) in this directory hold the reasoning for settled decisions.

Agents and contributors must not resolve an open decision by assumption. Record the question
here, name the owner, and wait for confirmation. See [the SDD workflow](../../specs/sdd-workflow.md).

## Status values

| Status | Meaning |
| --- | --- |
| Open | No decision yet. Work that depends on it is blocked. |
| Direction set | The direction is approved; concrete values or details still need confirmation. |
| Proposed | An option is recommended in an ADR or specification and is awaiting approval. |
| Decided | Approved. Recorded in an ADR where the reasoning matters. |
| Superseded | Replaced by a later decision. |

## Confirmation owners

| Owner | Responsibility |
| --- | --- |
| Legal | Agreement text, signature level, retention, erasure, lawful basis. |
| DPO | Data-protection assessment, data minimization, data-subject rights. |
| Platform owner | AWS accounts, regions, state backend, domains, WAF, cost. |
| Security owner | Authentication model, roles, key management. |
| Repository owner | Product scope, URLs, provider accounts, migration of existing records. |
| Maintainers | Tooling, CI, code conventions. |
| Organization admins | GitHub App registration and installation scope. |
| Design | Brand assets, fonts, final visual values. |

## Decisions made on 2026-10-05

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| A-01 | Specifications live under `specs/`; implementation and operational guides under `docs/`; ADRs under `docs/decisions/`. | Decided | [ADR-0001](0001-sdd-workflow-and-documentation-layout.md) |
| A-02 | `AGENTS.md` is rewritten for the platform, preserving governance, security, contribution, and repository-protection rules. | Decided | [ADR-0001](0001-sdd-workflow-and-documentation-layout.md) |
| A-03 | The repository stays public. | Decided | [ADR-0002](0002-public-cla-text-and-private-evidence.md) |
| A-04 | The public repository may hold canonical ICLA/CCLA text, version history, and signing instructions. Signed agreements, personal data, and audit evidence stay only in the private service. | Decided | [ADR-0002](0002-public-cla-text-and-private-evidence.md) |
| A-05 | `CONTRIBUTING.md` keeps describing CLA verification as manual until the platform is operational. | Decided | [ADR-0002](0002-public-cla-text-and-private-evidence.md) |
| A-06 | Refine the CI private-URL scan instead of changing the Go `internal/` layout. Documented local development URLs are allowed where appropriate. | Decided (details in D-01) | [ADR-0003](0003-refine-private-url-scan.md) |
| A-07 | Node.js 24 is pinned for local development and CI. | Decided | [ADR-0007](0007-pin-nodejs-24.md) |
| A-08 | The signature level is provider-configurable. The platform never claims a qualified signature unless the configured provider service is qualified. | Decided | [ADR-0004](0004-provider-neutral-signing-and-signature-level.md) |
| A-09 | S3 Object Lock retention is configurable. Retention periods and erasure obligations require a Legal decision. | Decided (values in D-10) | [ADR-0005](0005-configurable-object-lock-retention.md) |
| A-10 | CCLA authorization is never inferred from an email domain. | Decided | [ADR-0006](0006-explicit-ccla-authorization.md) |
| A-11 | Visual direction: deep navy or blue-green backgrounds, warm orange accents, white typography, rounded cards, subtle borders and gradients, strong hierarchy, consistent light and dark themes. | Decided (assets in D-12) | [Design tokens](../../specs/ux/design-tokens.md) |

## Decisions made at Phase 1 approval (2026-10-05)

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| D-01 | Private-URL scan matches private hostnames only in URL or host contexts, allows `localhost` and `127.0.0.1` development URLs, and supports a reviewed allowlist for exceptional cases. | Decided | [ADR-0003](0003-refine-private-url-scan.md) |
| D-04 | One AWS account per environment (`dev`, `staging`, `production`). Region `eu-west-1`; `eu-south-2` is the only permitted alternative. Account IDs are never committed. CloudTrail and account-level S3 Block Public Access are explicit platform-owner prerequisites, not assumed from an AWS Organizations baseline. | Decided (confirmed 2026-10-05) | [Architecture](../../specs/architecture.md#8-environments), [F-03](../../specs/features/F-03-terraform-foundation.md) |
| D-05 | Terraform S3 backend with S3-native locking (`use_lockfile = true`), one state bucket per account, created by a bootstrap stack. No DynamoDB lock table. | Decided | [Architecture](../../specs/architecture.md#9-infrastructure-as-code) |
| D-07 | npm workspaces, with Node.js 24 pinned for CI and local development. | Decided | [ADR-0007](0007-pin-nodejs-24.md) |
| D-17 | SQS queues, worker Lambdas, retries, visibility timeouts, and dead-letter queues. | Decided | [ADR-0008](0008-hexagonal-go-backend-on-lambda.md) |
| D-18 | Separate `core`, `audit`, and `ephemeral` DynamoDB tables. | Decided | [ADR-0009](0009-dynamodb-table-layout.md) |
| A-12 | The portal is hosted on CloudFront with a private S3 bucket and Origin Access Control. | Decided | [ADR-0010](0010-cloudfront-frontend-hosting.md) |

## Decisions made at F-01 approval (2026-10-06)

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| D-23 | Stylelint is deferred because its dependency `braces` has an unpatched high-severity advisory (GHSA-vfj7-8cjw-p6xm). Keeping it would weaken the `dependency-review` control. Token-only CSS validation moves to F-04, which adds Stylelint only when a patched `braces` is available or with a reviewed, time-limited `dependency-review` exception. Interim: ESLint forbids inline `style` attributes in `apps/web`. | Decided | [F-01](../../specs/features/F-01-monorepo-tooling.md), [design tokens](../../specs/ux/design-tokens.md#10-guardrails) |
| A-13 | Tool pins: Node.js 24, TypeScript 6.0.3, ESLint 9.39.5, Go 1.27.1, `golangci-lint` 2.14.0. Changing a pin requires a reviewed PR that updates F-01. | Decided | [F-01](../../specs/features/F-01-monorepo-tooling.md#implementation-plan) |
| A-14 | The Dev Container configuration is validated syntactically locally but was not built locally; the local failure is a host Docker DNS issue, and the configuration is not changed to work around it. The `Dev Container build` CI job builds the image and runs `make bootstrap` and `make check` before a PR can merge. | Decided | [F-02](../../specs/features/F-02-ci-pipelines.md) |

## Dependency update policy (2026-10-06)

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| A-15 | Dependabot ignores major upgrades of `eslint`, `@eslint/js`, `typescript`, and `@types/node` because they conflict with the A-13 pins: `eslint-plugin-jsx-a11y` 6.10.2 (latest) supports ESLint up to 9, `typescript-eslint` 8.71 supports TypeScript below 6.1, and `@types/node` must match the Node.js 24 runtime ([ADR-0007](0007-pin-nodejs-24.md)). Minor and patch updates continue. The pins are unchanged; upgrading any of these majors requires a reviewed PR that updates A-13 and F-01. Dependabot PRs #7 to #10 were closed under this policy. **Review point:** at the start of Phase 3 (before F-04), or earlier when `eslint-plugin-jsx-a11y` supports ESLint 10 or `typescript-eslint` supports TypeScript 7; `@types/node` is reviewed whenever the Node.js major in ADR-0007 changes. Owner: Maintainers. | Decided | [`.github/dependabot.yml`](../../.github/dependabot.yml), [F-01](../../specs/features/F-01-monorepo-tooling.md#implementation-plan) |
| A-16 | Amends A-15 for TypeScript (2026-10-07): Dependabot ignores every `typescript` version `>=6.1.0`, not only majors, because `typescript-eslint` 8.71 supports TypeScript below 6.1. Patch updates within 6.0 continue. The 6.0.3 pin is unchanged. **Review point:** when a `typescript-eslint` release supports TypeScript 6.1 or later, or at the start of Phase 3, whichever comes first. Owner: Maintainers. | Decided | [`.github/dependabot.yml`](../../.github/dependabot.yml) |

## Open decisions

The **Confirmation required from** column names who must confirm before the dependent work starts.
Decisions marked **Legal** or **Platform owner** cannot be settled by maintainers or agents.

| ID | Question | Status | Blocks | Confirmation required from | Recommendation |
| --- | --- | --- | --- | --- | --- |
| D-02 | Is `https://cla.developers.lamassu.cloud/` the production URL of this platform, or a separate static site? | Open | Phase 4 (OAuth redirect URIs), Phase 5 | Repository owner | None. Do not assume. |
| D-03 | Who supplies the canonical ICLA/CCLA text and version history, and in which languages? | Direction set | Phase 5 | **Legal** | Legal supplies the text; agents never draft it. |
| D-06 | Human authentication: GitHub OAuth App or GitHub App user-to-server tokens; session model. | Open | Phase 4 | Security owner | GitHub App user authorization, server-side sessions in DynamoDB, opaque HttpOnly cookie. |
| D-08 | GitHub App scope (organizations) and who registers it. | Open | Phase 8 | Organization admins | `lamassuiot` only at launch. |
| D-09 | Required signature level and evidence package per agreement type and jurisdiction. | Direction set | Phase 6 | **Legal** | Configuration per agreement type; default to the provider's standard electronic signature until Legal decides. |
| D-10 | Retention periods, Object Lock mode, erasure policy, data minimization. | Direction set | Phase 5 (storage), Phase 6, **any production deployment** | **Legal**, DPO | Governance mode outside production; production mode and period set by Legal. Production deployment stays blocked until confirmed. |
| D-11 | DocuSign account, environment, and integration-key ownership. | Open | Phase 6 | Repository owner | Developer (demo) account for dev and staging; production account owned by Lamassu. |
| D-12 | Brand assets, logos, and fonts that may be used. | Direction set | Phase 3 (final values) | Design, repository owner | Open-licensed fonts; no third-party logos. |
| D-13 | Role assignment for maintainers, legal administrators, and platform administrators. | Open | Phase 4 | Security owner | GitHub team membership mapped to roles, cached and re-checked per session. |
| D-14 | Must co-authors from `Co-authored-by` trailers be covered, and how are unverifiable trailer emails handled? | Open | Phase 8 | Maintainers, **Legal** | Require coverage; an unresolvable trailer yields a pending check, never a pass. |
| D-15 | Migration of existing manually verified signatories. | Open | Phases 6–7 | Repository owner, **Legal** | Import only with source evidence; otherwise require re-signing. |
| D-16 | Custom domains, certificates, and WAF per environment. | Open | Phase 9 | **Platform owner** | WAF and custom domains in staging and production. |
| D-19 | Email storage. | Direction set (technical default) | Phase 4 | **DPO**, Security owner | Technical default approved: persisted email lookup values are keyed HMACs only; raw email addresses are never persisted or logged. Key storage, rotation, and recovery are documented in [data protection](../../specs/security/data-protection.md#4-email-lookup-values-d-19). DPO confirmation pending. |
| D-20 | Lawful basis and privacy notice for contributor and signer data. | Open | Phase 4 (sign-in), Phase 6 | **Legal**, DPO | None. Legal decision. |
| D-21 | Supported languages for the portal and agreement text. | Open | Phase 3, Phase 5 | Repository owner, **Legal** | English first; others only with Legal-approved translations. |
| D-22 | When a new CLA version is published, must contributors with an active agreement on an older version re-sign, and by when? | Open (recommendation recorded) | Phase 5, Phase 8 | **Legal** | Substantive legal or scope changes require re-signing; purely editorial changes do not. Each published version records its change classification, set by Legal. Subject to Legal approval. |

## Summary of decisions awaiting confirmation

- **Legal:** D-03, D-09, D-10, D-14, D-15, D-20, D-21, D-22.
- **Platform owner:** D-16; creation of the three AWS accounts, verification of CloudTrail and
  account-level S3 Block Public Access, and execution of the Terraform bootstrap before any
  environment deployment.
- **DPO:** D-10, D-19, D-20.
- **Security owner:** D-06, D-13, D-19.
- **Repository owner:** D-02, D-11, D-15, D-21.
- **Organization admins:** D-08.
- **Design:** D-12.
