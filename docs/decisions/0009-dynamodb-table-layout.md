# ADR-0009: DynamoDB table layout

- Status: Accepted
- Date: 2026-10-05
- Decision-log entry: D-18

## Context

DynamoDB stores metadata and workflow state. Access patterns are known (see the
[domain model](../../specs/domain-model.md)). Audit events have different protection, retention,
and access requirements from operational data: they must be append-only and readable only by
authorized roles.

## Decision

Accepted: three tables per environment.

| Table | Contents | Key protections |
| --- | --- | --- |
| `core` | Contributors, agreements, signing sessions, organizations, authorized contributors, project policies, installations, CLA documents metadata. Single-table design. | Point-in-time recovery, deletion protection, KMS customer-managed key. |
| `audit` | Audit events only. | Append-only: IAM grants `PutItem` with a condition that the item does not exist, and denies `UpdateItem` and `DeleteItem`. Streams export to S3 evidence storage. |
| `ephemeral` | Webhook delivery records, idempotency keys, OAuth state, portal sessions. | TTL-based expiry. Short-lived; keys store hashes, not raw tokens. |

## Consequences

- IAM can isolate audit write and read permissions from operational permissions.
- TTL on the `ephemeral` table cannot accidentally expire core records.
- Single-table design in `core` requires careful key design and documented access patterns.
- Cross-table transactions are possible (`TransactWriteItems` spans tables in the same account and
  region), so state changes and their audit events can be written atomically.

## Alternatives considered

- **One table for everything.** Rejected: TTL and IAM boundaries are table-wide.
- **One table per entity.** Rejected: more tables and indexes to manage, and no access-pattern
  benefit.
