# Architecture Decision Records

This directory holds the [decision log](decision-log.md) and the architecture decision records
(ADRs) for the CLA Management Platform.

## When to write an ADR

Write an ADR when a decision:

- Changes or introduces an architectural component, data store, or integration.
- Changes an API contract, the data model, or a security assumption.
- Is hard to reverse or affects legal evidence or personal data.
- Resolves a decision in the decision log that needs recorded reasoning.

## Rules

- ADRs are numbered sequentially (`NNNN-short-title.md`) and never renumbered.
- An accepted ADR is not rewritten. To change it, add a new ADR that supersedes it and update the
  status of the old one to `Superseded by ADR-NNNN`.
- Every ADR links to the decision-log entry it resolves, if any.

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-sdd-workflow-and-documentation-layout.md) | SDD workflow and documentation layout | Accepted |
| [0002](0002-public-cla-text-and-private-evidence.md) | Public CLA text and private evidence | Accepted |
| [0003](0003-refine-private-url-scan.md) | Refine the private-URL scan | Accepted |
| [0004](0004-provider-neutral-signing-and-signature-level.md) | Provider-neutral signing and configurable signature level | Accepted |
| [0005](0005-configurable-object-lock-retention.md) | Configurable Object Lock retention | Accepted (values pending D-10) |
| [0006](0006-explicit-ccla-authorization.md) | Explicit CCLA authorization | Accepted |
| [0007](0007-pin-nodejs-24.md) | Pin Node.js 24 | Accepted |
| [0008](0008-hexagonal-go-backend-on-lambda.md) | Hexagonal Go backend on Lambda with asynchronous workers | Accepted |
| [0009](0009-dynamodb-table-layout.md) | DynamoDB table layout | Accepted |
| [0010](0010-cloudfront-frontend-hosting.md) | CloudFront frontend hosting | Accepted |

## ADR format

```markdown
# ADR-NNNN: Title

- Status: Proposed | Accepted | Superseded by ADR-NNNN
- Date: YYYY-MM-DD
- Decision-log entry: D-NN or A-NN

## Context

## Decision

## Consequences

## Alternatives considered
```
