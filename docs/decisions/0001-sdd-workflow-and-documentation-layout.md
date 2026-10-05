# ADR-0001: SDD workflow and documentation layout

- Status: Accepted
- Date: 2026-10-05
- Decision-log entries: A-01, A-02

## Context

The CLA Management Platform handles legal evidence and personal data. Uncontrolled changes to API
contracts, the data model, or security assumptions carry legal and privacy risk. The project uses
Specification-Driven Development (SDD) with VS Code agents, which needs a predictable place for
normative specifications and a single source of agent instructions.

`AGENTS.md` was inherited from the Lamassu repository template and described a reusable template
rather than this platform.

## Decision

- Normative specifications live under `specs/`: requirements, architecture, domain model, API,
  security, UX, feature specifications, and test scenarios.
- Implementation and operational guides live under `docs/`.
- Architecture decision records and the decision log live under `docs/decisions/`.
- The SDD workflow, approval gates, and the task-state file are defined in
  [`specs/sdd-workflow.md`](../../specs/sdd-workflow.md) and
  [`specs/STATE.md`](../../specs/STATE.md).
- `AGENTS.md` is rewritten for the platform. It keeps the inherited rules on information
  integrity, CODEOWNERS and review protection, commit signing, Conventional Commit PR titles, CLA
  handling, issue forms, Dependabot, Dev Containers, GitHub Projects, and legal boundaries.

## Consequences

- Code that contradicts `specs/` is a defect in either the code or the specification, and must be
  resolved explicitly.
- Every feature needs a specification and acceptance criteria before implementation starts.
- Documentation PRs become part of normal delivery, which adds review effort.

## Alternatives considered

- **Single `docs/` tree for everything.** Rejected: it blurs the line between normative
  specifications and explanatory guides.
- **Keep the template `AGENTS.md`.** Rejected: agents would receive instructions for a template,
  not for this platform.
