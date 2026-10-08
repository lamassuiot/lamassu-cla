# Task State

This file is the orchestrator's task-state record. Update it when a task starts, is blocked, or
completes. See [the SDD workflow](sdd-workflow.md).

Last updated: 2026-10-08

## Current position

- **Phase:** 1 complete (PR #4). Phase 2 in progress (remaining items blocked; see Blocked items).
  Phase 3 started on 2026-10-08 with the F-04 specification.
- **Phase 2:** F-02 unit 1 merged (PR #5). F-01 merged (PR #6), approved, and completed by
  PR #11 (`Dev Container build` CI job, now a required check). Dependency policy A-15 merged
  (PR #12). Biome migration (A-17, ADR-0011) merged (PR #14). F-02 unit 2 (backend workflow)
  merged (PR #15); unit 3 (frontend workflow) merged (PR #17). Unit 4 (OpenAPI lint; generated
  client deferred by A-18) merged (PR #18). Unit 6 (CodeQL and required-check list) merged
  (PR #19). F-03 steps 1 to 4 and F-02 unit 5 (Terraform workflow) merged (PR #20); no
  deployment. A-19 approved conditionally on 2026-10-08: Checkov replaces Trivy (ADR-0012), merged
  in PR #22; the scan runs without network access in CI (PR #23). The three Checkov dependency
  alerts were dismissed in Dependabot as tolerable risk, referencing ADR-0012, on 2026-10-08.
  A-19, ADR-0012, and T-19 remain proposed.
- **Ruleset:** the 10 Backend, Frontend, OpenAPI, and CodeQL checks were added to the `main`
  ruleset on 2026-10-07; the 4 Terraform checks were added on 2026-10-08 (20 required checks).
- **Deployment gate:** No branch deploys infrastructure or creates AWS resources until the
  platform-owner prerequisites in [F-03](features/F-03-terraform-foundation.md#preconditions) are
  met. Production additionally requires Legal confirmation of D-10.
- **Next gate:** approval of the [F-04 specification](features/F-04-design-tokens-and-themes.md)
  before its implementation. Phase 2: Security-owner and platform-owner review of the final A-19,
  ADR-0012, and T-19 wording; approval of D-24 (central state account). F-03 step 5 (deployment
  workflow) stays blocked until the platform-owner prerequisites are met (see Blocked items).

## Approvals

| Date | Approver | Scope |
| --- | --- | --- |
| 2026-10-05 | Repository owner | [Discovery report](discovery-report.md), Phase 1 deliverable order, decisions A-01 to A-11. |
| 2026-10-05 | Repository owner | Phase 1 artifacts; D-01, D-04, D-05, D-07, D-17 (ADR-0008), D-18 (ADR-0009), A-12 (ADR-0010); D-19 technical default subject to DPO; D-22 recommendation recorded subject to Legal; start of Phase 2 tooling and CI work. |
| 2026-10-05 | Repository owner | F-01, F-02, and F-03 (implementation, no deployment); F-02 unit 1 implementation; region `eu-west-1` with `eu-south-2` as sole alternative; CloudTrail and account-level Block Public Access as explicit platform-owner prerequisites; production blocked until D-10. |
| 2026-10-06 | Repository owner | F-01 implementation; D-23 (Stylelint deferred, token-only CSS validation moved to F-04); A-13 tool pins (Node.js 24, TypeScript 6.0.3, ESLint 9.39.5, Go 1.27.1, `golangci-lint` 2.14.0); A-14 (local Dev Container build failure accepted as a host Docker networking issue; CI must build the Dev Container before merge). |
| 2026-10-06 | Repository owner | A-15 (Dependabot defers major upgrades of ESLint, `@eslint/js`, TypeScript, and `@types/node`; pins unchanged; review at the start of Phase 3); `Dev Container build` added as a required check in the `main` ruleset. |
| 2026-10-07 | Repository owner | A-16 (Dependabot defers TypeScript `>=6.1.0` until `typescript-eslint` supports it; 6.0 patch updates continue). |
| 2026-10-07 | Repository owner | A-17 (migrate from ESLint and Prettier to Biome if the required controls can be implemented; TypeScript stays at 6.0.3). Implemented in PR #14. |
| 2026-10-07 | Repository owner | A-18 (F-02 unit 4: Redocly 2.60.0 OpenAPI lint in CI; generated client and drift check deferred, not removed; TypeScript pin unchanged, no npm overrides, no Orval). |
| 2026-10-07 | Repository owner | Add the 10 Backend, Frontend, OpenAPI, and CodeQL checks to the `main` ruleset; implement F-03 with F-02 unit 5 without AWS deployment. |
| 2026-10-08 | Repository owner | A-19 approved conditionally: Checkov replaces Trivy, evaluated, pinned, and integrity-verified; runs without AWS credentials; `terraform validate`, TFLint, and plan tests remain independent controls; A-19 updated only after Checkov detects the insecure fixtures. Every failed Checkov check fails the scan, because severities are unavailable offline; exceptions only as justified inline skips. Hash-locked pip install with a temporary dependency-review exception for GHSA-9w56-46f6-3qhx, GHSA-89v8-rhwq-hf77, and GHSA-wj6h-64fc-37mp (ADR-0012). Fix CKV_AWS_300; skip CKV_AWS_144, CKV_AWS_18, and CKV2_AWS_62 on the state bucket pending D-24. T-19 describes the target state architecture; D-24 is added as a platform-owner prerequisite. A-19, ADR-0012, and T-19 stay proposed until Checkov is validated and the Security owner and platform owner review the final wording. Terraform work may continue with plan-only, mocked-provider tests only. |
| 2026-10-08 | Repository owner | Start Phase 3 while the remaining Phase 2 items stay blocked; write the F-04 specification (specification-only PR). Final visual values remain pending D-12. |

## Phases

| Phase | Name | Status | Blocked by |
| --- | --- | --- | --- |
| 1 | Discovery and specification foundation | Done (PR #4) | — |
| 2 | Repository and tooling foundation | In progress | Deployment only: platform-owner prerequisites (F-03); production also D-10. |
| 3 | Frontend shell and design system | In progress (F-04 specification) | Final visual values: D-12. |
| 4 | Authentication and user profile | Not started | Phase 2; D-02, D-06, D-13, D-19 (DPO), D-20. |
| 5 | Public CLA catalogue | Not started | Phase 4; D-03, D-10, D-21. |
| 6 | ICLA workflow | Not started | Phase 5; D-09, D-11, D-15. |
| 7 | CCLA workflow | Not started | Phase 6; D-15. |
| 8 | GitHub App validation | Not started | Phase 6; D-08, D-14. |
| 9 | Operations and hardening | Not started | Phases 2–8; D-16. |

## Phase 1 deliverables

| ID | Deliverable | Status |
| --- | --- | --- |
| P1-1 | Orchestration foundation: `AGENTS.md`, [SDD workflow](sdd-workflow.md), this file, [decision log](../docs/decisions/decision-log.md) | Approved |
| P1-2 | [Product requirements](requirements.md) | Approved |
| P1-3 | [Architecture specification](architecture.md) and ADRs | Approved; ADR-0008 to ADR-0010 accepted |
| P1-4 | [Domain model](domain-model.md) | Approved |
| P1-5 | [OpenAPI structure](api/openapi.yaml) and [API conventions](api/conventions.md) | Approved |
| P1-6 | [Security specification](security/security.md), [threat model](security/threat-model.md), [data-protection notes](security/data-protection.md) | Approved; Legal, DPO, and Security-owner items remain open |
| P1-7 | [Design-token specification](ux/design-tokens.md) | Approved; final values pending D-12 |

## Planned features

Feature specifications are written at the start of the phase that implements them.

| ID | Feature | Phase | Depends on | Status |
| --- | --- | --- | --- | --- |
| F-01 | [Monorepo and tooling foundation](features/F-01-monorepo-tooling.md) | 2 | Phase 1 | Done (PRs #6, #11); Biome migration (A-17) merged (PR #14) |
| F-02 | [CI pipelines for backend, frontend, OpenAPI, and Terraform](features/F-02-ci-pipelines.md) | 2 | F-01 | Unit 1 merged (PR #5); Dev Container job merged (PR #11); unit 2 merged (PR #15); unit 3 merged (PR #17); unit 4 merged (PR #18; generated client deferred, A-18); unit 6 merged (PR #19); unit 5 merged (PR #20) |
| F-03 | [Terraform foundation and environments](features/F-03-terraform-foundation.md) | 2 | F-01 | Steps 1–4 merged (PR #20; no deployment); Checkov scan merged (PRs #22, #23; ADR-0012 proposed); step 5 blocked; A-19, T-19, and D-24 pending |
| F-04 | [Design tokens and themes](features/F-04-design-tokens-and-themes.md), including token-only CSS validation (D-23) | 3 | F-01 | Specification in review |
| F-05 | Component library and preview | 3 | F-04 | Not started |
| F-06 | App shell and static screens | 3 | F-05 | Not started |
| F-07 | GitHub sign-in and sessions | 4 | F-02, F-03, D-06 | Not started |
| F-08 | Roles and authorization middleware | 4 | F-07, D-13 | Not started |
| F-09 | Public CLA catalogue | 5 | F-08, D-03 | Not started |
| F-10 | Signing-provider port and DocuSign adapter | 6 | F-03, D-09, D-11 | Not started |
| F-11 | ICLA signing workflow | 6 | F-09, F-10 | Not started |
| F-12 | Organizations and CCLA signing | 7 | F-11 | Not started |
| F-13 | Authorized contributors, review, revocation, revalidation | 7 | F-12 | Not started |
| F-14 | GitHub App webhook intake | 8 | F-03, D-08 | Not started |
| F-15 | CLA coverage evaluation and Check Runs | 8 | F-11, F-13, F-14, D-14 | Not started |
| F-16 | Project policies | 8 | F-14 | Not started |
| F-17 | Audit-log view and exports | 9 | F-11 | Not started |
| F-18 | Monitoring, alarms, and production readiness | 9 | All | Not started |

## Contradictions

| ID | Contradiction | Proposed resolution | Status |
| --- | --- | --- | --- |
| C-01 | `infra/bootstrap` and [F-03](features/F-03-terraform-foundation.md#main-flow) create the state bucket and KMS key in each workload account. They use one role for state access and deployment, with state keys under `<environment>/` (D-05). [T-19](security/threat-model.md#3-threats-and-mitigations) and D-24 describe a dedicated Infrastructure/Tooling state account, `lamassu-cla/<environment>/<stack>/terraform.tfstate` keys, and separate state and deployment roles. | After D-24 is approved, rework `infra/bootstrap`, F-03, and the [architecture](architecture.md#9-infrastructure-as-code) in a separate pull request, with plan-only, mocked-provider tests. Until then, the current bootstrap must not be run. | Open; awaiting D-24 |

## Blocked items

See the [decision log](../docs/decisions/decision-log.md) for owners. Terraform implementation may
continue with plan-only, mocked-provider tests.

- **F-03 step 5 and any AWS operation.** No bootstrap, no plan against AWS, and no apply in any
  environment until the platform owner confirms the Infrastructure/Tooling account (D-24), the
  workload accounts, the Control Tower controls, and the GitHub deployment environments, together
  with the other [F-03 prerequisites](features/F-03-terraform-foundation.md#preconditions).
  Production also requires D-10.
