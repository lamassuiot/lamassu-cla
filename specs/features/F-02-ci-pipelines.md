# F-02: CI pipelines

| Field | Value |
| --- | --- |
| Status | Approved; all units implemented (unit 5 merged in PR #20; unit 4 partially: generated client deferred, A-18) |
| Phase | 2 |
| Owner | Repository owner |
| Depends on | F-01 (units 2 onward); D-01; A-18; [ADR-0003](../../docs/decisions/0003-refine-private-url-scan.md), [ADR-0007](../../docs/decisions/0007-pin-nodejs-24.md) |
| Approved by | Repository owner, 2026-10-05 (all units; unit 1 implementation approved); 2026-10-07 (A-18: unit 4 without the generated client) |

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
| Generated API client out of date | Regenerate and `git diff --exit-code` | Job fails | Diff shown (deferred, A-18) |
| OpenAPI contract violates a lint rule | `redocly lint` with `redocly.yaml` | Job fails | Rule, location, and code frame |
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
  *(Unit 2)*
- **AC-02-6** Frontend workflow: `npm ci`, type check, Biome (`biome ci`), the TypeScript
  directive check, Vitest, and build. Token-only CSS validation is added with F-04 (D-23, A-17).
  *(Unit 3)*
- **AC-02-7** OpenAPI workflow: Redocly lint with a committed configuration *(Unit 4)* and a
  generated-client drift check *(**deferred** by A-18; still required before any feature consumes
  API operations through `packages/api-client`)*.
- **AC-02-8** Terraform workflow: `fmt -check`, `validate` without a backend, TFLint, and a
  configuration security scan, for every environment and module. *(Unit 5)*
- **AC-02-9** CodeQL analyzes Go and TypeScript on pull requests and weekly. *(Unit 6)*
- **AC-02-10** The required-check names are documented for the ruleset administrator. *(Unit 6)*
- **AC-02-11** The `CI` workflow's `Dev Container build` job builds `.devcontainer/` with a pinned
  Dev Container CLI and `--frozen-lockfile`, runs `make bootstrap`, and runs `make check` in the
  container. It is a required check (A-14). *(F-01 follow-up)*

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F02-T1 | `scripts/check-private-urls.test.sh` (14 cases). | Script | AC-02-1 to AC-02-3 |
| F02-T2 | `actionlint` and workflow schema validation. | CI | All workflows |
| F02-T3 | A deliberately stale generated client fails the drift check (verified once during implementation). **Deferred (A-18).** | CI | AC-02-7 |
| F02-T4 | The `Dev Container build` job passes on the pull request that adds it. | CI | AC-02-11 |
| F02-T5 | The `Backend` workflow's `Go checks` and `Go vulnerability scan` jobs pass on the pull request that adds them. | CI | AC-02-5 |
| F02-T6 | The `Frontend` workflow's `Frontend checks` job passes on the pull request that adds it. | CI | AC-02-6 |
| F02-T7 | `redocly lint` passes on the contract and fails on a deliberately invalid contract (verified once during implementation); the `OpenAPI lint` job passes on the pull request that adds it. | CI | AC-02-7 (lint) |
| F02-T8 | The `CodeQL Go` and `CodeQL TypeScript` jobs complete and upload results on the pull request that adds them. | CI | AC-02-9 |
| F02-T9 | Every check name in the [required-check list](#required-checks) matches a job name reported on a pull request. | CI | AC-02-10 |
| F02-T10 | The `Terraform` workflow's `Terraform checks`, `TFLint`, and `Terraform security scan` jobs pass on the pull request that adds them; Trivy fails on a deliberately weakened bucket module (verified once during implementation). | CI | AC-02-8 |

## Observability requirements

Job summaries list the checks run. No secrets or environment identifiers appear in logs.

## Rollback considerations

Revert the workflow change. If a required check is renamed, update the ruleset in the same
maintenance window to avoid blocking merges.

## Implementation plan

1. **Implemented:** refine the private-URL scan into a tested script with a reviewed allowlist;
   pin Node.js through `.nvmrc`.
   - **F-01 follow-up:** `Dev Container build` job in `CI` (AC-02-11), approved 2026-10-06.
2. **Implemented:** backend workflow (`.github/workflows/backend.yml`).
3. **Implemented:** frontend workflow (`.github/workflows/frontend.yml`).
4. **Implemented (partially):** OpenAPI workflow (`.github/workflows/openapi.yml`) and Redocly
   configuration (`redocly.yaml`).
   - **Deferred (A-18):** generated API client in `packages/api-client` and its drift check.
5. **Implemented:** Terraform workflow (`.github/workflows/terraform.yml`), delivered with F-03.
6. **Implemented:** CodeQL workflow (`.github/workflows/codeql.yml`) and the
   [required-check list](#required-checks).

Implementation notes (unit 2):

- Path filtering runs in a `Detect backend changes` job, not in `on.pull_request.paths`: a
  workflow skipped by a path filter leaves required checks pending, while a job skipped by `if`
  reports success. Backend jobs run when `services/api/`, the `Makefile`, or the workflow
  changes, on every push to `main`, and whenever detection does not report `false`.
- Jobs call the `make` targets, so CI and local runs share the Go toolchain and tool pins.
  `govulncheck` is pinned to v1.8.0 in the `Makefile` (`make vuln-go`); it needs network access
  and is not part of `make check`.
- The weekly schedule (Mondays 06:00 UTC) runs only `Go vulnerability scan`.

Implementation notes (unit 3):

- `Frontend checks` uses the same in-job path filtering as unit 2. Because Biome checks the whole
  repository, it runs when `apps/`, `packages/`, `.biome/`, the TypeScript directive scripts,
  `.npmrc`, `.nvmrc`, `.gitignore`, the workflow, or any JavaScript, TypeScript, JSON, CSS, or
  HTML file changes.
- Steps call the root npm scripts, as `make lint-web`, `make test-web`, and `make build-web` do.

Implementation notes (unit 4):

- Redocly CLI is pinned to 2.60.0 as a root development dependency. `npm run lint:openapi`
  (`make lint-openapi`, part of `make lint`) runs `redocly lint --lint-config=error`, so an
  invalid `redocly.yaml` also fails.
- `redocly.yaml` extends `recommended-strict` and sets `telemetry: off`, which applies to local,
  Dev Container, and CI runs.
- `.redocly.lint-ignore.yaml` lists only the 16 `no-unused-components` findings for shared
  components declared before any operation uses them. New unused components and every other rule
  still fail. Remove an entry when an operation starts using the component.
- `OpenAPI lint` uses the same in-job path filtering as units 2 and 3. It runs when
  `specs/api/`, the Redocly files, npm manifests, `.npmrc`, `.nvmrc`, or the workflow change.
- The generated client and drift check are deferred by A-18: the evaluated candidates conflict
  with TypeScript 6.0.3, introduce audit vulnerabilities, or need further evaluation. The
  TypeScript pin is unchanged and no npm `overrides` are added.

Implementation notes (unit 6):

- CodeQL runs as two jobs instead of a matrix, because a skipped matrix job reports an unexpanded
  name that never satisfies a required check. `CodeQL Go` builds `services/api` manually;
  `CodeQL TypeScript` uses `build-mode: none`. Both use the default query suite.
- Path filtering follows units 2 to 4: analysis runs when Go, JavaScript, or TypeScript sources,
  workspace directories, npm manifests, TypeScript configuration, or the workflow change. Pushes
  to `main` and the weekly schedule (Mondays 06:30 UTC) always analyze.
- Only the analysis jobs get `security-events: write`. `github/codeql-action` is pinned to the
  v4.38.2 commit already used by the Scorecard workflow.
- Code scanning default setup is not configured for the repository. Enabling it would conflict
  with this workflow; keep it disabled.

Implementation notes (unit 5):

- `Terraform checks` runs `make fmt-terraform` and `make test-terraform` (init without a backend,
  `validate`, and plan-only `terraform test` with mocked providers) for every module and root.
  `TFLint` runs `make lint-terraform`; `Terraform security scan` runs `make scan-terraform`.
- No job assumes an AWS role or reads credentials. Plans against real accounts belong to the
  deployment workflow (F-03 step 5), which is not implemented.
- Terraform 1.16.5 and TFLint 0.64.0 match the Dev Container. `hashicorp/setup-terraform` and
  `terraform-linters/setup-tflint` are pinned by commit SHA. Trivy is downloaded by the `Makefile`
  and verified against a pinned SHA-256; no third-party scanning action is used (A-19).
- Path filtering follows the other workflows: the jobs run when `infra/`, `.tflint.hcl`, the
  `Makefile`, or the workflow change.

### Required checks

Configuring the `main` ruleset is a manual administrator task; this list is the source of truth.
Names are the job names that GitHub reports on pull requests.

| Check | Workflow | Required since |
| --- | --- | --- |
| `Validate PR title format` | Lint PR Title | Instantiation |
| `Validate repository template` | CI | Instantiation |
| `GitHub Actions security lint` | CI | Instantiation |
| `Secret scan` | CI | Instantiation |
| `Dependency review` | CI | Instantiation |
| `Dev Container build` | CI | A-14 |
| `Detect backend changes` | Backend | 2026-10-07 |
| `Go checks` | Backend | 2026-10-07 |
| `Go vulnerability scan` | Backend | 2026-10-07 |
| `Detect frontend changes` | Frontend | 2026-10-07 |
| `Frontend checks` | Frontend | 2026-10-07 |
| `Detect OpenAPI changes` | OpenAPI | 2026-10-07 |
| `OpenAPI lint` | OpenAPI | 2026-10-07 |
| `Detect CodeQL changes` | CodeQL | 2026-10-07 |
| `CodeQL Go` | CodeQL | 2026-10-07 |
| `CodeQL TypeScript` | CodeQL | 2026-10-07 |
| `Detect Terraform changes` | Terraform | 2026-10-08 |
| `Terraform checks` | Terraform | 2026-10-08 |
| `TFLint` | Terraform | 2026-10-08 |
| `Terraform security scan` | Terraform | 2026-10-08 |

Notes for the administrator:

- Jobs skipped by path filtering report success, so every check above can be required without
  blocking documentation-only pull requests. Requiring the `Detect ...` jobs makes a detection
  failure visible.
- `Go vulnerability scan` queries the Go vulnerability database. A newly published advisory for the
  pinned Go toolchain or a module blocks backend pull requests until the pin is updated; this is
  intended.
- A successful `CodeQL ...` job means the analysis completed, not that no alerts exist. Blocking
  merges on code scanning alerts needs the ruleset's code scanning rule, which is a separate
  decision.
- `OpenSSF Scorecard`, `Release`, and checks from external GitHub Apps are not required.
- Add a check to the ruleset only after the workflow that reports it is on `main`; otherwise
  open pull requests based on an older `main` cannot satisfy it.

## Open questions

- Generated API client tool and drift check: deferred by A-18 with a review point; owner
  Maintainers.
