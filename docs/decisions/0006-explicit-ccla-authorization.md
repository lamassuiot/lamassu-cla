# ADR-0006: Explicit CCLA authorization

- Status: Accepted
- Date: 2026-10-05
- Decision-log entry: A-10

## Context

A CCLA covers contributions made on behalf of an organization. Matching a contributor's email
domain to an organization's domain is easy to spoof (personal accounts on shared domains,
contractors, former employees, lookalike domains) and does not prove that the organization
authorized the contribution.

## Decision

- An organization's CCLA is signed by an **authorized representative** whose authority is stated
  in the signing flow and reviewed by a maintainer or legal administrator before the CCLA becomes
  active.
- A contributor is covered by a CCLA only when they appear on the organization's **authorized
  contributor list**, identified by GitHub user ID, and the list entry was added by the
  organization's representative or an authorized platform role.
- Email domain matching is never used to grant coverage, authority, or membership. It may be shown
  only as an informational hint to reviewers.
- Coverage is revoked when the CCLA is suspended or revoked, or the contributor is removed from the
  list. Authorized contributors are revalidated periodically.

## Consequences

- Organizations must actively maintain their contributor lists.
- Contributors may see a pending CLA check until their organization adds them.
- Automated tests must include a case where a contributor's verified email domain matches an
  organization but the contributor is not covered.

## Alternatives considered

- **Domain-based auto-enrollment.** Rejected: it does not establish authorization.
- **GitHub organization membership as proof.** Rejected as sole proof: GitHub organization
  membership does not prove legal authority. It may become an optional, reviewed input later.
