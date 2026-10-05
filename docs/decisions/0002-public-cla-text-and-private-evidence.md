# ADR-0002: Public CLA text and private evidence

- Status: Accepted
- Date: 2026-10-05
- Decision-log entries: A-03, A-04, A-05; related open decision D-03

## Context

The repository is public. The platform publishes CLA documents that anyone may read, and it stores
signed agreements, contributor identities, corporate authorization evidence, and audit records
that are private.

The inherited `AGENTS.md` prohibited copying the full ICLA or CCLA into repositories created from
the template. That rule targets consumer repositories, which must only link to the central CLA
portal. This repository is the CLA platform itself.

`CONTRIBUTING.md` documents a manual verification process for contributors to this repository.

## Decision

- The repository stays public.
- This repository **may** contain the canonical ICLA and CCLA text, their version history, and
  signing instructions, under a dedicated directory defined in the Phase 5 feature specification.
- The canonical text is supplied and approved by Legal (D-03). Agents and contributors never draft,
  paraphrase, or edit agreement text.
- Signed agreements, personal data, corporate authorization evidence, and audit evidence are stored
  **only** in the private service (private S3 buckets and DynamoDB tables). They never enter this
  repository, its issues, its pull requests, public project boards, CI logs, or build artifacts.
- `CONTRIBUTING.md` keeps describing CLA verification as manual until the platform is operational
  in production. Switching to automated verification is a separate, reviewed change.
- Other Lamassu repositories still must not copy the agreement text; they link to the portal.

## Consequences

- The published text in Git and the published document served by the platform must be identical.
  The SHA-256 hash of the published document is the link between them.
- Any change to the agreement text is a Legal-approved PR, and a published version is immutable.
- Test fixtures must use synthetic identities and synthetic documents only.
- The threat model, security specification, and operational guides are public and must not
  disclose account identifiers, resource names, or incident details.

## Alternatives considered

- **Keep agreement text outside the repository.** Rejected by the repository owner: public version
  history in Git supports transparency and review.
- **Make the repository private.** Rejected: the platform is open source.
