# ADR-0013: Pin Node.js 26

- Status: Accepted
- Date: 2026-11-27
- Decision-log entry: A-28
- Supersedes: [ADR-0007](0007-pin-nodejs-24.md)

## Context

ADR-0007 pinned Node.js 24 as the only supported release line and made upgrading it a separate,
reviewed change. Dependabot PR #77 moved `playwright` to 1.65.1, and the matching official
Playwright image (`mcr.microsoft.com/playwright:v1.65.1-noble`, A-23) ships Node.js 26 instead of
Node.js 24. No Node.js 24 variant of that image is published. With `engine-strict`, `npm ci` in
the browser-test job rejected the image's runtime, and the lockstep rule of A-23 (image tag equals
`playwright` version) could not be kept on Node.js 24.

Node.js 26 is the current LTS release line.

## Decision

- Node.js 26 is the only supported Node.js release line for local development and CI.
- `.nvmrc` is `26`; the root `package.json` `engines.node` is `>=26 <27`; `@types/node` follows
  the 26 line; the Dev Container installs Node.js 26.13.0 or later from the 26 line; every
  workflow that sets up Node.js keeps `node-version-file: .nvmrc`.
- The Playwright image and the `playwright` package stay in lockstep (A-23). A future image whose
  bundled Node.js leaves the pinned line triggers this review again rather than a silent change.
- Upgrading the Node.js release line remains a separate, reviewed change.

## Consequences

- Contributors on Node.js 24 must switch, for example with `nvm use`.
- `make check`, the browser tests, and the 36 visual baselines were run on Node.js 26 before the
  change merged; no baseline changed, because screenshots depend on Chromium, not Node.js.
- The lockfile is regenerated only for `@types/node`; application dependencies are unchanged.

## Alternatives considered

- **Keep Node.js 24 and pin `playwright` back to 1.64.1.** Rejected by the repository owner: it
  freezes the browser-test toolchain on an image whose runtime is no longer the LTS line.
- **Install Node.js 24 inside the Playwright container job.** Rejected: CI would run a different
  runtime from the image used for local `make test-visual`, contrary to ADR-0007's goal of one
  runtime everywhere.
