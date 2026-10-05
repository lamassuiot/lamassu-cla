# ADR-0005: Configurable Object Lock retention

- Status: Accepted (values pending D-10)
- Date: 2026-10-05
- Decision-log entries: A-09, D-10

## Context

Signed agreements and evidence packages must be protected against replacement and deletion. S3
Object Lock provides write-once-read-many (WORM) protection, but:

- Object Lock must be enabled when the bucket is created.
- In **compliance** mode, no user, including the root user, can delete a locked object version or
  shorten its retention before it expires.
- In **governance** mode, users with a specific permission can bypass the lock.
- Personal data in locked objects cannot be erased during the retention period, which may conflict
  with data-erasure obligations.

The required retention periods and the balance between evidence preservation and erasure rights
are legal decisions.

## Decision

- Evidence buckets are created with S3 Object Lock enabled, versioning enabled, KMS encryption, and
  S3 Block Public Access.
- The Object Lock **mode** and **default retention period** are Terraform variables per environment,
  with validation. They have no production default: production `apply` fails until Legal-approved
  values are set.
- Non-production environments use governance mode with a short retention period, so test data can
  be cleaned up.
- Legal holds can be applied per object by authorized roles, independently of retention periods.
- Personal data is minimized in locked objects. Lookup data such as email addresses is kept in
  DynamoDB, where it can be erased or pseudonymized, rather than in locked object metadata.
- The legal decision required is documented in
  [the data-protection notes](../../specs/security/data-protection.md).

## Consequences

- Production infrastructure cannot be deployed before Legal confirms D-10.
- Choosing compliance mode makes the retention period irreversible for every stored object.
- Erasure requests may need to be answered with "retained under legal obligation" for locked
  evidence. Legal must approve this wording.

## Alternatives considered

- **Hard-code compliance mode with a fixed period.** Rejected: the period is a legal decision.
- **No Object Lock; rely on IAM deny policies.** Rejected: administrators could remove the
  policies, which weakens evidence integrity.
