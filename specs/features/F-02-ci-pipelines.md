# F-02: CI pipelines

| Field | Value |
| --- | --- |
| Status | Approved; unit 1 implemented |
| Phase | 2 |
| Owner | Repository owner |
| Depends on | F-01 (units 2 onward); D-01; [ADR-0003](../../docs/decisions/0003-refine-private-url-scan.md), [ADR-0007](../../docs/decisions/0007-pin-nodejs-24.md) |
| Approved by | Repository owner, 2026-10-05 (all units; unit 1 implementation approved) |

## Problem statement

Pull requests must run every relevant check for the backend, frontend, OpenAPI contract, and
Terraform, with stable check names that rulesets can require. The inherited private-URL scan
blocks the planned Go layout and local development documentation.

## User roles

Contributors, maintainers, and repository administrators (who configure required checks).

## Preconditions

- F-01 manifests exist (units 2 onward).
- Actions are pinned by commit SHA, as in the existing workflows.

## Main flow

1. A pull request triggers the workflows whose path filters match the change.
2. Each job runs with read-only `contents` permission unless it needs more.
3. Results report under stable job names that are listed as required checks.

## Alternative flows

- Changes limited to documentation run only the governance workflow (`CI`).
- Scheduled runs (CodeQL, `govulncheck`) run weekly on `main`.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Private hostname in a URL | `scripts/check-private-urls.sh` | Job fails | File, line, and allowlist guidance |
| Allowlist entry without justification | Same script | Job fails (exit 2) | Allowlist line named |
| Generated API client out of date | Regenerate and `git diff --exit-code` | Job fails | Diff shown |
| Terraform not formatted | `terraform fmt -check` | Job fails | Files listed |

## Security considerations

- Workflows default to `permissions: contents: read`. `pull_request_target` remains limited to
  the existing title-lint workflow.
- No workflow on pull requests assumes AWS roles. Terraform plans against real accounts run only in
  protected GitHub environments (F-03).
- Terraform plan output is never posted to public pull requests.
- The private-URL scan reads tracked and untracked, non-ignored files.

## API changes

None. The OpenAPI workflow validates the contract.

## Data-model changes

None.

## UI changes

None.

## Acceptance criteria

- **AC-02-1** The private-URL scan fails on URLs whose host contains an `internal`, `intranet`, or
  `corp` label, and passes on Go `internal/` import paths, dotted identifiers, and `localhost` or
  `127.0.0.1` URLs. *(Unit 1)*
- **AC-02-2** Allowlist entries are fixed strings with a justification comment directly above; an
  entry without one fails the scan. *(Unit 1)*
- **AC-02-3** The scan has automated tests that run in CI before the scan. *(Unit 1)*
- **AC-02-4** Every workflow that sets up Node.js uses `node-version-file: .nvmrc`, which contains
  `24`. *(Unit 1)*
- **AC-02-5** Backend workflow: `go vet`, `golangci-lint`, `go test -race`, and `govulncheck`.
- **AC-02-6** Frontend workflow: `npm ci`, type check, ESLint, Prettier check, Vitest, and build.
  Stylelint is added with F-04 (D-23).
- **AC-02-7** OpenAPI workflow: Redocly lint with a committed configuration and a generated-client
  drift check.
- **AC-02-8** Terraform workflow: `fmt -check`, `validate` without a backend, TFLint, and a
  configuration security scan, for every environment and module.
- **AC-02-9** CodeQL analyzes Go and TypeScript on pull requests and weekly.
- **AC-02-10** The required-check names are documented for the ruleset administrator.
- **AC-02-11** The `CI` workflow's `Dev Container build` job builds `.devcontainer/` with a pinned
  Dev Container CLI and `--frozen-lockfile`, runs `make bootstrap`, and runs `make check` in the
  container. It is a required check (A-14). *(F-01 follow-up)*

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F02-T1 | `scripts/check-private-urls.test.sh` (14 cases). | Script | AC-02-1 to AC-02-3 |
| F02-T2 | `actionlint` and workflow schema validation. | CI | All workflows |
| F02-T3 | A deliberately stale generated client fails the drift check (verified once during implementation). | CI | AC-02-7 |
| F02-T4 | The `Dev Container build` job passes on the pull request that adds it. | CI | AC-02-11 |

## Observability requirements

Job summaries list the checks run. No secrets or environment identifiers appear in logs.

## Rollback considerations

Revert the workflow change. If a required check is renamed, update the ruleset in the same
maintenance window to avoid blocking merges.

## Implementation plan

1. **Implemented:** refine the private-URL scan into a tested script with a reviewed allowlist;
   pin Node.js through `.nvmrc`.
   - **F-01 follow-up:** `Dev Container build` job in `CI` (AC-02-11), approved 2026-10-06.
2. Backend workflow.
3. Frontend workflow.
4. OpenAPI workflow and Redocly configuration.
5. Terraform workflow (with F-03).
6. CodeQL workflow and required-check documentation.

## Open questions

None.
