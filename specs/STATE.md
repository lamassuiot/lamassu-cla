# Task State

This file is the orchestrator's task-state record. Update it when a task starts, is blocked, or
completes. See [the SDD workflow](sdd-workflow.md).

Last updated: 2026-10-05

## Current position

- **Phase:** 1 approved; Phase 1 pull request open from `docs/sdd-discovery-report`.
- **Phase 2:** tooling and CI work may proceed. Application implementation (Phases 3 onward) must
  not begin until the Phase 1 pull request is reviewed and merged.
- **Next gate:** Approval of the Phase 2 feature specifications F-01, F-02, and F-03.

## Approvals

| Date | Approver | Scope |
| --- | --- | --- |
| 2026-10-05 | Repository owner | [Discovery report](discovery-report.md), Phase 1 deliverable order, decisions A-01 to A-11. |
| 2026-10-05 | Repository owner | Phase 1 artifacts; D-01, D-04, D-05, D-07, D-17 (ADR-0008), D-18 (ADR-0009), A-12 (ADR-0010); D-19 technical default subject to DPO; D-22 recommendation recorded subject to Legal; start of Phase 2 tooling and CI work. |

## Phases

| Phase | Name | Status | Blocked by |
| --- | --- | --- | --- |
| 1 | Discovery and specification foundation | Approved; PR open | Merge of the Phase 1 PR. |
| 2 | Repository and tooling foundation | In progress | Approval of F-01 to F-03 before scaffolding. Terraform apply: platform-owner provisioning of accounts and the bootstrap stack. |
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

| ID | Feature | Phase | Depends on |
| --- | --- | --- | --- |
| F-01 | Monorepo and tooling foundation | 2 | Phase 1 |
| F-02 | CI pipelines for backend, frontend, OpenAPI, and Terraform | 2 | F-01 |
| F-03 | Terraform foundation and environments | 2 | F-01 |
| F-04 | Design tokens and themes | 3 | F-01 |
| F-05 | Component library and preview | 3 | F-04 |
| F-06 | App shell and static screens | 3 | F-05 |
| F-07 | GitHub sign-in and sessions | 4 | F-02, F-03, D-06 |
| F-08 | Roles and authorization middleware | 4 | F-07, D-13 |
| F-09 | Public CLA catalogue | 5 | F-08, D-03 |
| F-10 | Signing-provider port and DocuSign adapter | 6 | F-03, D-09, D-11 |
| F-11 | ICLA signing workflow | 6 | F-09, F-10 |
| F-12 | Organizations and CCLA signing | 7 | F-11 |
| F-13 | Authorized contributors, review, revocation, revalidation | 7 | F-12 |
| F-14 | GitHub App webhook intake | 8 | F-03, D-08 |
| F-15 | CLA coverage evaluation and Check Runs | 8 | F-11, F-13, F-14, D-14 |
| F-16 | Project policies | 8 | F-14 |
| F-17 | Audit-log view and exports | 9 | F-11 |
| F-18 | Monitoring, alarms, and production readiness | 9 | All |

## Contradictions

None recorded.

## Blocked items

See the [decision log](../docs/decisions/decision-log.md) for owners. Phase 2 has no open decision
blockers. Applying Terraform requires the platform owner to provision the AWS accounts and run the
bootstrap stack.
