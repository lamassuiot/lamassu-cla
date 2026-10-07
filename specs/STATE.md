# Task State

This file is the orchestrator's task-state record. Update it when a task starts, is blocked, or
completes. See [the SDD workflow](sdd-workflow.md).

Last updated: 2026-10-07

## Current position

- **Phase:** 1 complete (PR #4). Phase 2 in progress.
- **Phase 2:** F-02 unit 1 merged (PR #5). F-01 merged (PR #6), approved, and completed by
  PR #11 (`Dev Container build` CI job, now a required check). Dependency policy A-15 merged
  (PR #12). Biome migration (A-17, ADR-0011) merged (PR #14). F-02 unit 2 (backend workflow)
  merged (PR #15); unit 3 (frontend workflow) merged (PR #17). Unit 4 (OpenAPI lint; generated
  client deferred by A-18) merged (PR #18). Unit 6 (CodeQL and required-check list) merged
  (PR #19). F-03 steps 1 to 4 with F-02 unit 5 (Terraform workflow) implemented on
  `feat/terraform-foundation-main`, in review; no deployment. A-19 proposed.
- **Ruleset:** the 10 Backend, Frontend, OpenAPI, and CodeQL checks were added to the `main`
  ruleset on 2026-10-07 (16 required checks). The 4 Terraform checks are added after unit 5
  merges.
- **Deployment gate:** No branch deploys infrastructure or creates AWS resources until the
  platform-owner prerequisites in [F-03](features/F-03-terraform-foundation.md#preconditions) are
  met. Production additionally requires Legal confirmation of D-10.
- **Next gate:** Approval of F-03 steps 1 to 4, F-02 unit 5, and A-19; then F-03 step 5
  (deployment workflow), which stays disabled until the platform-owner prerequisites are met.

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

## Phases

| Phase | Name | Status | Blocked by |
| --- | --- | --- | --- |
| 1 | Discovery and specification foundation | Done (PR #4) | — |
| 2 | Repository and tooling foundation | In progress | Deployment only: platform-owner prerequisites (F-03); production also D-10. |
| 3 | Frontend shell and design system | Not started | Phase 2; final visual values D-12. |
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
| F-02 | [CI pipelines for backend, frontend, OpenAPI, and Terraform](features/F-02-ci-pipelines.md) | 2 | F-01 | Unit 1 merged (PR #5); Dev Container job merged (PR #11); unit 2 merged (PR #15); unit 3 merged (PR #17); unit 4 merged (PR #18; generated client deferred, A-18); unit 6 merged (PR #19); unit 5 implemented with F-03, in review |
| F-03 | [Terraform foundation and environments](features/F-03-terraform-foundation.md) | 2 | F-01 | Steps 1–4 implemented, awaiting approval (no deployment); step 5 not started |
| F-04 | Design tokens and themes, including token-only CSS validation (D-23) | 3 | F-01 | Not started |
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

None recorded.

## Blocked items

See the [decision log](../docs/decisions/decision-log.md) for owners. Phase 2 implementation has no
open blockers. Any environment deployment requires the platform-owner prerequisites in
[F-03](features/F-03-terraform-foundation.md#preconditions); production also requires D-10.
