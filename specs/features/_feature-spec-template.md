# F-NN: Feature name

| Field | Value |
| --- | --- |
| Status | Draft, In review, Approved, In progress, Done |
| Phase | Phase number from [`STATE.md`](../STATE.md) |
| Owner | GitHub handle of the person responsible |
| Depends on | Feature IDs and decision IDs |
| Approved by | Name and date, recorded when approved |

## Problem statement

What problem does this feature solve, and for whom?

## User roles

Roles from the [security specification](../security/security.md) that interact with this feature.

## Preconditions

What must be true before the main flow starts.

## Main flow

Numbered steps.

## Alternative flows

Named variations of the main flow.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Example | How it is detected | What the system does | What the user sees |

## Security considerations

Trust boundaries, authentication, authorization, input validation, secrets, personal data, and
threat-model entries affected.

## API changes

New or changed operations in [`openapi.yaml`](../api/openapi.yaml). State whether any change is
breaking.

## Data-model changes

New or changed entities, keys, indexes, and state transitions in
[`domain-model.md`](../domain-model.md).

## UI changes

Screens, components, and states (loading, empty, error, permission denied). Only existing design
tokens and components may be used.

## Acceptance criteria

Use testable statements, for example in Given/When/Then form. Number them `AC-NN-1`, `AC-NN-2`.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| Example | What is tested | Unit, integration, contract, end-to-end | Acceptance criteria IDs |

## Observability requirements

Structured log events, metrics, alarms, and audit events.

## Rollback considerations

How to roll back the deployment and any data changes. Note irreversible steps.

## Implementation plan

Bounded units, each suitable for one pull request, with dependencies.

## Open questions

Link each to a decision-log entry.
