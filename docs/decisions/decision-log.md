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
| D-05 | Terraform S3 backend with S3-native locking (`use_lockfile = true`), one state bucket per account, created by a bootstrap stack. No DynamoDB lock table. **Proposed amendment (D-24):** state moves to a dedicated Infrastructure/Tooling account. | Decided | [Architecture](../../specs/architecture.md#9-infrastructure-as-code) |
| D-07 | npm workspaces, with Node.js 24 pinned for CI and local development. | Decided | [ADR-0007](0007-pin-nodejs-24.md) |
| D-17 | SQS queues, worker Lambdas, retries, visibility timeouts, and dead-letter queues. | Decided | [ADR-0008](0008-hexagonal-go-backend-on-lambda.md) |
| D-18 | Separate `core`, `audit`, and `ephemeral` DynamoDB tables. | Decided | [ADR-0009](0009-dynamodb-table-layout.md) |
| A-12 | The portal is hosted on CloudFront with a private S3 bucket and Origin Access Control. | Decided | [ADR-0010](0010-cloudfront-frontend-hosting.md) |

## Decisions made at F-01 approval (2026-10-06)

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| D-23 | Stylelint is deferred because its dependency `braces` has an unpatched high-severity advisory (GHSA-vfj7-8cjw-p6xm). Keeping it would weaken the `dependency-review` control. Token-only CSS validation moves to F-04, which adds Stylelint only when a patched `braces` is available or with a reviewed, time-limited `dependency-review` exception. Interim: ESLint forbids inline `style` attributes in `apps/web`. **Amended by A-17:** the interim inline-style control is now Biome; F-04 first evaluates Biome CSS rules and GritQL plugins for token-only validation and uses Stylelint only if they cannot express it. | Decided (amended by A-17) | [F-01](../../specs/features/F-01-monorepo-tooling.md), [design tokens](../../specs/ux/design-tokens.md#10-guardrails) |
| A-13 | Tool pins: Node.js 24, TypeScript 6.0.3, ESLint 9.39.5, Go 1.27.1, `golangci-lint` 2.14.0. Changing a pin requires a reviewed PR that updates F-01. **Amended by A-17:** the ESLint 9.39.5 pin is superseded by Biome 2.5.15; the other pins are unchanged. | Decided (amended by A-17) | [F-01](../../specs/features/F-01-monorepo-tooling.md#implementation-plan) |
| A-14 | The Dev Container configuration is validated syntactically locally but was not built locally; the local failure is a host Docker DNS issue, and the configuration is not changed to work around it. The `Dev Container build` CI job builds the image and runs `make bootstrap` and `make check` before a PR can merge. | Decided | [F-02](../../specs/features/F-02-ci-pipelines.md) |

## Dependency update policy (2026-10-06)

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| A-15 | Dependabot ignores major upgrades of `eslint`, `@eslint/js`, `typescript`, and `@types/node` because they conflict with the A-13 pins: `eslint-plugin-jsx-a11y` 6.10.2 (latest) supports ESLint up to 9, `typescript-eslint` 8.71 supports TypeScript below 6.1, and `@types/node` must match the Node.js 24 runtime ([ADR-0007](0007-pin-nodejs-24.md)). Minor and patch updates continue. The pins are unchanged; upgrading any of these majors requires a reviewed PR that updates A-13 and F-01. Dependabot PRs #7 to #10 were closed under this policy. **Review point:** at the start of Phase 3 (before F-04), or earlier when `eslint-plugin-jsx-a11y` supports ESLint 10 or `typescript-eslint` supports TypeScript 7; `@types/node` is reviewed whenever the Node.js major in ADR-0007 changes. Owner: Maintainers. **Amended by A-17:** the `eslint` and `@eslint/js` entries are removed with those packages; the `@types/node` rule and its review point remain. | Decided (amended by A-17) | [`.github/dependabot.yml`](../../.github/dependabot.yml), [F-01](../../specs/features/F-01-monorepo-tooling.md#implementation-plan) |
| A-16 | Amends A-15 for TypeScript (2026-10-07): Dependabot ignores every `typescript` version `>=6.1.0`, not only majors, because `typescript-eslint` 8.71 supports TypeScript below 6.1. Patch updates within 6.0 continue. The 6.0.3 pin is unchanged. **Review point:** when a `typescript-eslint` release supports TypeScript 6.1 or later, or at the start of Phase 3, whichever comes first. Owner: Maintainers. **Amended by A-17:** `typescript-eslint` is removed, so the ignore rule now only protects the A-13 pin; it stays until a separate decision changes the TypeScript pin, reviewed at the start of Phase 3. | Decided (amended by A-17) | [`.github/dependabot.yml`](../../.github/dependabot.yml) |

## Decisions made on 2026-10-07

| ID | Decision | Status | Record |
| --- | --- | --- | --- |
| A-17 | Biome 2.5.15 (exact pin) replaces ESLint and Prettier for JavaScript, TypeScript, JSX, JSON, CSS, and HTML formatting and linting. `tsc --noEmit` and Vitest are unchanged. Inline `style` attributes in `apps/web` fail through `nursery/noInlineStyles` and a GritQL plugin; rules without a Biome equivalent use GritQL plugins, `tsc`, or `scripts/check-ts-directives.sh`. Supersedes the ESLint pin in A-13 and the ESLint entries in A-15; amends the rationale of A-16 and the interim control in D-23. TypeScript stays at 6.0.3. | Decided | [ADR-0011](0011-biome-for-javascript-and-typescript.md) |
| A-18 | F-02 unit 4 adds Redocly CLI 2.60.0 (exact pin) and lints `specs/api/openapi.yaml` in CI with the committed `redocly.yaml` (`recommended-strict`, telemetry off). The 16 shared components that the Phase 1 contract declares before any operation uses them are listed in `.redocly.lint-ignore.yaml`. **The generated API client and its drift check (part of AC-02-7) are deferred, not removed.** Reason: the evaluated candidates either conflict with the TypeScript 6.0.3 pin (A-13) (`openapi-typescript` 7.13.0 requires TypeScript 5), introduce audit vulnerabilities that would fail `dependency-review` (`@hey-api/openapi-ts` 0.97.0 and 0.99.0: high-severity `js-yaml` advisories), or require further evaluation (`orval` 8.40.0). The TypeScript pin is unchanged, no npm `overrides` are added, and Orval is not introduced. **Review point:** at the start of Phase 3, or earlier when a candidate installs with the pinned TypeScript without overrides and passes `npm audit`; the generated client and drift check must exist before any feature consumes API operations through `packages/api-client`. Owner: Maintainers. | Decided (deferral) | [F-02](../../specs/features/F-02-ci-pipelines.md#implementation-plan) |
| A-19 | Terraform tooling for F-03 and F-02 unit 5: AWS provider `>= 6.67.0, < 7.0.0`, locked to 6.68.0 in each root module; Terraform 1.16.5 and TFLint 0.64.0 (as in the Dev Container); TFLint AWS ruleset 0.49.0. **Amended 2026-10-08 (conditional approval):** Checkov 3.3.26 replaces Trivy. It is installed from a hash-locked requirements file for Python 3.12 (`--require-hashes`, wheels only), runs with an empty environment and no AWS credentials, and fails on every failed check, because severities are unavailable offline; exceptions are justified inline skips. Before each scan, `make test-terraform-scan` must detect the insecure fixtures in `scripts/testdata/checkov/` (verified 2026-10-08 in the Dev Container). `terraform validate`, TFLint, and plan tests remain independent controls. Dependency review has a temporary exception for three advisories in Checkov's dependencies. Dependabot updates the root modules and the Checkov lock. Changing a pin requires a reviewed PR that updates F-03. | Proposed (pending Security-owner and platform-owner review of the final wording) | [ADR-0012](0012-checkov-terraform-scan.md), [F-03](../../specs/features/F-03-terraform-foundation.md#implementation-plan) |

## Phase 3 decision reviews (2026-10-08)

The start of Phase 3 is the review point set by ADR-0012, A-15, A-16, and A-18. The original
entries above are kept unchanged as decision history; each review below records its evidence and
proposed outcome. Evaluations ran outside the repository (temporary worktrees, temporary npm
projects, and the pinned Python 3.12 image) and changed no pin.

| ID | Review, evidence, and outcome | Status | Record |
| --- | --- | --- | --- |
| A-20 | **Review of the ADR-0012 dependency-review exception.** *Evidence:* (1) Checkov 3.3.26 (2026-10-07) is still the latest release, and Checkov's `main` branch (`setup.py`, `Pipfile`) still pins `asteval==1.0.6` exactly; upstream issue bridgecrewio/checkov#7504 and PR #7517 are open. No reproducible, supported Checkov installation can use a patched `asteval` (1.0.9 or 1.0.10). (2) A hash-pinned override to `asteval` 1.0.10 works functionally (fixture test 13 passed; the `infra/` scan returns the same 35 results as 1.0.6), but `pip check` fails (`checkov 3.3.26 has requirement asteval==1.0.6`), so it is an unsupported combination and would require removing the `pip check` gate. (3) Checkov creates its interpreter with `Interpreter(symtable=SAFE_EVAL_DICT, use_numpy=False, minimal=True)` (`safe_eval_functions.py`). With the locked 1.0.6, the GHSA-9w56-46f6-3qhx chain fails with `NameError` (no numpy symbols), and the GHSA-89v8-rhwq-hf77 exception classes are not reachable (`SystemExit`, `KeyboardInterrupt`, `BaseException` undefined; `raise` is rejected with an ordinary `ValueError`). (4) GHSA-wj6h-64fc-37mp (`ecdsa` 0.19.2) still has no patched version (`>= 0`). Checkov imports `ecdsa` only in `common/external_checks/verification` (`VerifyingKey`, signature verification of external checks); `verify_and_register` is a no-op without public-key paths, and the scan passes none. The advisory is a timing attack on signing with a private key, which does not occur. (5) CI isolation was verified in the `main` run of PR #23 (Terraform run 37748012899, job `Terraform security scan`): token `Contents: read`, `Metadata: read`; hash-locked install with `pip check` passing; fixture test (13 passed) and scan (32 passed, 0 failed, 3 skipped) both ran under `sudo unshare --net -- setpriv --reuid=1001`. The pre-scan probe only proves that one connection failed; a stronger check (only the `lo` interface is visible in the namespace) is proposed in a separate PR. *Outcome:* keep the exception unchanged, limited to exactly the three GHSA identifiers; do not override `asteval`. *Next review:* when a Checkov release allows `asteval>=1.0.9` (the Dependabot `pip` PR), when `ecdsa` publishes a fix or Checkov uses it for signing, when the scan starts using external checks or keys, or at the start of Phase 4, whichever comes first. | Proposed (pending Security-owner review) | [ADR-0012](0012-checkov-terraform-scan.md#review-2026-10-08-a-20) |
| A-21 | **Review of A-15 and A-16 (TypeScript and `@types/node` deferrals).** *Evidence:* (1) The A-15 and A-16 reasons for the TypeScript rule were `typescript-eslint` peer ranges; A-17 removed ESLint and `typescript-eslint`. (2) No package in `package-lock.json` declares `typescript` as a dependency or peer; only the root `devDependencies` uses it. (3) No TypeScript 6.1 exists: after 6.0.3 (2026-04-16) the next release is 7.0.2 (2026-07-08, `latest`), so the `>=6.1.0` rule blocks only TypeScript 7. (4) In a temporary worktree of `main` at 57a9409, TypeScript 7.0.2 (exact) passed `npm run typecheck`, `biome ci`, Vitest in every workspace, and `npm run build`; `npm audit` found 0 vulnerabilities; a planted type error failed with TS2322. The upgrade adds about 20 platform-specific `@typescript/typescript-*` packages to the lockfile. *Outcome:* remove Dependabot's `typescript >=6.1.0` ignore rule so that upgrades are proposed as PRs and tested by CI. TypeScript stays pinned at 6.0.3 (A-13); an upgrade merges only in a separate PR that passes the frontend checks and is reviewed, and that PR updates A-13 and F-01. Keep the `@types/node` semver-major ignore aligned with Node.js 24 ([ADR-0007](0007-pin-nodejs-24.md)); review it when the Node.js major changes. | Proposed (pending repository-owner approval) | [`.github/dependabot.yml`](../../.github/dependabot.yml) |
| A-22 | **Scheduled review of A-18 (generated API client).** *Evidence* (Node.js 24.21, TypeScript 6.0.3 exact, no `overrides`, no `--legacy-peer-deps`, contract `specs/api/openapi.yaml` with 1 operation and 4 schemas): (1) `openapi-typescript` 7.13.0 with `openapi-fetch` 0.17.0 does not install: peer `typescript@^5.x` conflicts with 6.0.3. It generates types only; the runtime client is `openapi-fetch`. (2) `@hey-api/openapi-ts` 0.99.0 installs, but `npm audit` reports 4 high-severity findings through `js-yaml` 4.2.0 (GHSA-52cp-r559-cp3m, GHSA-5p4m-2wfm-xmqj, GHSA-2883-xcg3-v3hh) in `@hey-api/json-schema-ref-parser` 1.4.4. (3) `orval` 8.40.0 installs without conflicts (81 packages, `npm audit` 0 vulnerabilities; licenses MIT, ISC, Apache-2.0, BSD-2-Clause, BSD-3-Clause, Python-2.0, and MIT OR CC0-1.0). `esbuild` 0.28.2's install script is not run, and `esbuild` works from its platform package. With `client: 'fetch'`, Orval generates both types (schemas, responses, parameters) and a usable runtime function (`getHealth()` calling `GET /v1/health` with typed 200, 429, and 503 results). The output type-checks under `tsconfig.base.json` with the DOM library, a mocked-`fetch` test returned typed 200 and 503 results, generation is byte-identical when repeated without network access, and a contract edit changes the output, so a regenerate-and-compare drift check is reliable. The output has no Biome lint findings but needs `biome format` (65 lines). *Gaps for implementation:* the generated function uses the global `fetch` with relative URLs and no session or CSRF handling ([security](../../specs/security/security.md)), so a reviewed custom fetch wrapper (Orval `mutator`) is required; response bodies are typed, not validated at runtime. *Outcome:* Orval 8.40.0 is the first candidate that meets the A-18 installation and audit criteria. Propose adopting it (exact pin, `packages/api-client` development dependency) to complete AC-02-7 in a separate implementation PR. The A-18 deferral continues until that PR merges; the generated client and drift check remain required before any feature consumes API operations through `packages/api-client`. | Proposed (pending repository-owner approval) | [F-02](../../specs/features/F-02-ci-pipelines.md#implementation-plan) |

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
| D-24 | Centralized Terraform state. Store the state of every environment in a dedicated Infrastructure/Tooling account: never in application workload accounts, and never in the Control Tower management, Log Archive, or Audit accounts. Use one state bucket and one customer-managed KMS key per environment (`dev`, `staging`, `production`). Use state keys `lamassu-cla/<environment>/<stack>/terraform.tfstate`, with S3-native locking through the corresponding `.tflock` object. Use separate roles for backend state access and for workload-account deployment, each restricted to its environment and state prefix. **Platform-owner prerequisite:** the tooling account must exist and be confirmed by the platform owner; its existence is not assumed. Amends D-05 and the F-03 bootstrap design (contradiction C-01 in `specs/STATE.md`). | Proposed | F-03 step 5; any bootstrap, plan against AWS, or apply | **Platform owner**, Security owner | Target state in [T-19](../../specs/security/threat-model.md#3-threats-and-mitigations). Accept the residual risk that a tooling-account compromise exposes several environment states. |

## Summary of decisions awaiting confirmation

- **Legal:** D-03, D-09, D-10, D-14, D-15, D-20, D-21, D-22.
- **Platform owner:** D-16, D-24; creation and confirmation of the Infrastructure/Tooling account
  (D-24) and the three workload accounts, confirmation of the Control Tower controls, verification
  of CloudTrail and account-level S3 Block Public Access, and execution of the Terraform bootstrap
  before any environment deployment. Review of the final A-19, ADR-0012, and T-19 wording.
- **DPO:** D-10, D-19, D-20.
- **Security owner:** D-06, D-13, D-19, D-24; review of A-19, ADR-0012 (including the temporary
  dependency-review exception), and T-19; review of A-20 (Phase 3 review of the exception).
- **Repository owner:** D-02, D-11, D-15, D-21; approval of A-21 and A-22.
- **Organization admins:** D-08.
- **Design:** D-12.
