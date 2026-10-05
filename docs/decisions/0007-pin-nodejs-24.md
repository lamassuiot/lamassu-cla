# ADR-0007: Pin Node.js 24

- Status: Accepted
- Date: 2026-10-05
- Decision-log entry: A-07

## Context

The CI workflow uses Node.js 24. A maintainer workstation runs Node.js 23, which is a
non-LTS release line. Different runtimes can produce different lockfiles, builds, and test
results.

## Decision

- Node.js 24 is the only supported Node.js release line for local development and CI.
- Phase 2 adds `.nvmrc` with `24`, an `engines.node` constraint in the root `package.json`, and
  `node-version-file: .nvmrc` in every workflow that sets up Node.js.
- Upgrading the Node.js release line is a separate, reviewed change.

## Consequences

- Contributors on other Node.js versions must switch, for example with `nvm use`.
- The engine constraint makes installs on unsupported versions fail early.

## Alternatives considered

- **Track the latest Node.js release.** Rejected: odd-numbered releases are not LTS.
