# SDD Workflow and Orchestration Rules

This document defines how humans and VS Code agents deliver work on the CLA Management Platform
using Specification-Driven Development (SDD). It is normative. [`AGENTS.md`](../AGENTS.md) requires
every agent to follow it.

## 1. Workflow

```text
requirements
  → architecture
  → feature specification
  → acceptance criteria
  → implementation plan
  → implementation
  → tests
  → security review
  → documentation update
  → pull request
```

Each arrow is a gate. Work does not move to the next stage until the current stage's exit criteria
are met.

| Stage | Output | Exit criteria |
| --- | --- | --- |
| Requirements | [`requirements.md`](requirements.md) | Reviewed and approved by the repository owner. |
| Architecture | [`architecture.md`](architecture.md), [`domain-model.md`](domain-model.md), ADRs | Reviewed; proposed ADRs accepted or explicitly deferred. |
| Feature specification | `specs/features/F-NN-name.md` from the [template](features/_feature-spec-template.md) | All template sections complete; open questions logged. |
| Acceptance criteria | In the feature specification | Every criterion is testable and mapped to a test scenario. |
| Implementation plan | In the feature specification | Bounded units, each small enough for one reviewable PR. |
| **Approval gate** | Approval recorded in [`STATE.md`](STATE.md) | Explicit approval from the repository owner or a maintainer. |
| Implementation | Code, Terraform, OpenAPI changes | Matches the specification; no unapproved contract changes. |
| Tests | Automated tests | All new and existing tests pass locally and in CI. |
| Security review | Checklist in the PR | [Security specification](security/security.md) checklist completed. |
| Documentation update | Updated `specs/`, `docs/`, ADRs | Specifications describe what was built. |
| Pull request | PR with Conventional Commit title | CI passes; Code Owner review requested. |

## 2. Orchestrator responsibilities

The orchestrator is the agent (or person) coordinating a phase or feature. It must:

1. Read [`AGENTS.md`](../AGENTS.md), this document, [`STATE.md`](STATE.md), and the
   [decision log](../docs/decisions/decision-log.md) before doing any work.
2. Read the specifications relevant to the task.
3. Inspect the repository before proposing changes. Do not rely on memory of earlier sessions.
4. Update [`STATE.md`](STATE.md) when a task starts, is blocked, or completes.
5. Break work into bounded features and units, and record their dependencies.
6. Refuse to start implementation of a feature that has no approved specification and acceptance
   criteria.
7. Ask for approval at each feature boundary and record it.
8. Run the relevant tests and validations after each implementation unit.
9. Update specifications and ADRs in the same PR as the implementation change.
10. Detect contradictions between code and specifications and report them (see section 4).
11. Record every unresolved question in the decision log with an owner.
12. Keep each PR focused and reviewable.

## 3. Change control

These changes are **never** made silently. Each one requires an updated specification, an ADR when
the change is significant, and explicit approval before merging:

| Change | Required artifacts |
| --- | --- |
| API contract (paths, schemas, status codes, auth requirements) | Updated `specs/api/openapi.yaml`, feature specification, changelog entry. Breaking changes use `!` in the PR title. |
| Data model (entities, keys, indexes, states, invariants) | Updated [`domain-model.md`](domain-model.md), migration and rollback notes. |
| Security assumption (trust boundary, authentication, authorization, secret handling, retention) | Updated [security specification](security/security.md) and [threat model](security/threat-model.md); security-owner review. |
| Legal behavior (agreement text, signature level, retention, coverage rules) | Legal confirmation recorded in the decision log. |
| Design tokens (new color, spacing, or other value) | Updated [design-token specification](ux/design-tokens.md) with justification. |
| New infrastructure component | ADR and architecture update. |

## 4. Contradiction handling

When code, specifications, or ADRs disagree:

1. Stop work on the affected unit.
2. Record the contradiction in [`STATE.md`](STATE.md) under **Contradictions** with file
   references.
3. Propose a resolution: fix the code, or update the specification with justification.
4. Wait for approval if the resolution changes behavior, a contract, the data model, or a security
   assumption.

## 5. Approval gates

Approval is required:

- Before starting each phase.
- Before implementing each feature (after its specification and acceptance criteria exist).
- Before accepting any proposed ADR.
- Before any change listed in section 3.

Record approvals in [`STATE.md`](STATE.md) with the date and approver. An approval in chat is valid
only once recorded.

## 6. Feature specifications

Feature specifications live in `specs/features/` and are named `F-NN-short-name.md`. Copy the
[feature specification template](features/_feature-spec-template.md). Every section is mandatory;
write "None" with a reason when a section does not apply.

## 7. Pull requests

- One feature unit per PR. Specification-only PRs are encouraged before implementation PRs.
- PR titles follow Conventional Commits (see [`CONTRIBUTING.md`](../CONTRIBUTING.md)).
- The PR description links the feature specification and lists the acceptance criteria covered.
- PRs never include secrets, personal data, signed agreements, or legal evidence.

## 8. Definition of done

A feature is done only when:

- Its specification exists and acceptance criteria are explicit.
- API and data-model changes are documented.
- Frontend behavior is specified.
- Security implications are reviewed.
- Tests are implemented and pass.
- Observability (logs, metrics, alarms) is included.
- Documentation is updated.
- Terraform changes are validated.
- CI passes.
- No design-system guardrail is violated.
- The implementation is ready for code review.
