# F-01: Monorepo and tooling foundation

| Field | Value |
| --- | --- |
| Status | Done (PR #6; Dev Container CI validation in PR #11); lint and format tooling migrated to Biome (A-17) |
| Phase | 2 |
| Owner | Repository owner |
| Depends on | Phase 1; D-07; [ADR-0007](../../docs/decisions/0007-pin-nodejs-24.md), [ADR-0008](../../docs/decisions/0008-hexagonal-go-backend-on-lambda.md), [ADR-0011](../../docs/decisions/0011-biome-for-javascript-and-typescript.md) |
| Approved by | Repository owner, 2026-10-05 (specification); 2026-10-06 (implementation, D-23, A-13, A-14); 2026-10-07 (Biome migration, A-17) |

## Problem statement

The repository has no build tooling. Contributors and CI need one reproducible way to install,
lint, test, and build the Go backend, the frontend workspaces, and shared packages before any
application feature is implemented.

## User roles

Contributors and maintainers (developer experience). No end-user roles are affected.

## Preconditions

- Phase 1 specifications merged (PR #4).
- Node.js 24 (`.nvmrc`), Go, and npm available locally.

## Main flow

1. A contributor clones the repository and runs `nvm use` and `make bootstrap`.
2. `make lint`, `make test`, and `make build` run every workspace's checks.
3. The same targets run in CI (F-02).

## Alternative flows

- A contributor works on one area only: workspace-scoped commands (`npm run test -w apps/web`,
  `go test ./...` in `services/api`) behave the same as the `make` targets.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Unsupported Node.js version | `engines` with `engine-strict` | Install fails | Message naming Node.js 24 |
| Unsupported Go version | `go` directive in `go.mod` | Build fails or toolchain download | Go version error |
| Layer rule violation | `depguard` in `golangci-lint` | Lint fails | Rule and import named |
| Inline `style` attribute in the portal | Biome `nursery/noInlineStyles` and the `.biome/plugins/no-inline-style-attribute.grit` plugin | Lint fails | Message pointing to design-system tokens |
| Triple-slash reference, `@ts-nocheck`, or undescribed `@ts-expect-error` | `scripts/check-ts-directives.sh` | Lint fails | File, line, and reason |

Token-only CSS validation is deferred to F-04 (D-23, amended by A-17).

## Security considerations

- Lockfiles (`package-lock.json`, `go.sum`) are committed; CI uses `npm ci`.
- `.npmrc` sets `engine-strict=true` and does not configure registries or tokens.
- No secrets, environment identifiers, or real personal data in configuration or fixtures.
- Dependabot entries for `gomod` and `npm` are added in the same PR as the manifests.

## API changes

None.

## Data-model changes

None.

## UI changes

`apps/web` contains only the Vite entry point and an empty application shell placeholder. Screens
are built in Phase 3.

## Acceptance criteria

- **AC-01-1** Given Node.js 24, when `npm ci` runs at the root, then all workspaces install from the
  committed lockfile. Given another major version, installation fails.
- **AC-01-2** The root `package.json` is private and declares the workspaces `apps/*` and
  `packages/*`.
- **AC-01-3** `services/api` is a Go module named `github.com/lamassuiot/lamassu-cla/services/api`
  with the layer directories from the [architecture](../architecture.md#31-layers).
- **AC-01-4** `golangci-lint` fails when `internal/domain` imports another layer or any AWS, GitHub,
  or provider SDK, and when a layer other than `internal/adapters` imports an SDK.
- **AC-01-5** `make lint`, `make test`, and `make build` succeed on a clean clone.
- **AC-01-6** TypeScript runs in `strict` mode in every workspace; Biome formats and lints all
  JavaScript, TypeScript, JSX, JSON, CSS, and HTML files (`biome ci`), with the controls listed in
  [ADR-0011](../../docs/decisions/0011-biome-for-javascript-and-typescript.md).
- **AC-01-7** `.github/dependabot.yml` includes `gomod` for `/services/api` and `npm` for `/`, with
  weekly schedules, a limit of 5, grouping, and Conventional Commit prefixes.
- **AC-01-8** `docs/development/local-setup.md` documents prerequisites and every `make` target.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F01-T1 | A placeholder unit test per workspace and Go package runs and passes. | Unit | AC-01-5 |
| F01-T2 | `scripts/check-go-layering.test.sh`: 9 forbidden imports (cross-layer, adapters, I/O in `domain`) fail with a `depguard` finding; 4 allowed imports pass. | Lint | AC-01-4 |
| F01-T3 | CI installs with `npm ci` on Node.js 24. | CI | AC-01-1 |
| F01-T4 | The `Dev Container build` CI job builds the Dev Container from the frozen lockfile, runs `make bootstrap`, and passes `make check` (F-02, A-14). | CI | AC-01-5 |
| F01-T5 | `scripts/check-ts-directives.test.sh` (10 cases). Every rule mapped in ADR-0011, the inline-style plugin, and the two TypeScript plugins fire on synthetic violations and not on valid code (verified once during the A-17 migration). | Script, Lint | AC-01-6 |

## Observability requirements

None at runtime. Structured logging is introduced with the first handler (Phase 4).

## Rollback considerations

Tooling-only changes; revert the PR. No data or infrastructure is affected.

## Implementation plan

1. Root npm workspace, `.npmrc`, Biome (originally ESLint and Prettier; see A-17), TypeScript base
   configuration, Vitest.
2. Workspace skeletons: `apps/web` (Vite, React, TypeScript), `packages/design-system`,
   `packages/api-client`, `packages/shared-types`.
3. Go module skeleton with layer packages, `golangci-lint` configuration with `depguard`, and one
   compiling `cmd/api` entry point that returns no business behavior.
4. Root `Makefile`, `.editorconfig` additions for Go (tabs) and Terraform, Dependabot entries.
5. Local development guide.
6. Dev Container, added to CODEOWNERS in the same PR (per `AGENTS.md`).

Implementation notes:

- TypeScript is pinned to 6.0.3 (A-13). The original reason, the `typescript-eslint` peer range,
  no longer applies after A-17; changing the pin needs a separate decision.
- Biome is pinned to 2.5.15 (A-17), replacing the ESLint 9.39.5 pin.
- Go 1.27.1 is pinned in `go.mod`; the `Makefile` sets `GOTOOLCHAIN` so older local Go installs
  download it automatically. `golangci-lint` v2.14.0 is installed into `.tools/bin`.
- Stylelint was removed after `npm audit` reported an unpatched `braces` advisory (D-23).
- The pins above are accepted (A-13). Changing one requires a reviewed PR that updates this
  section.
- Dependabot ignores major upgrades of `@types/node` and every `typescript` version `>=6.1.0`
  until the A-15 and A-16 review points; other minor and patch updates continue.
- The Dev Container configuration was validated syntactically (`devcontainer read-configuration`
  and Feature resolution) but not built locally: the host's Docker daemon could not resolve
  package hosts. The configuration is not changed for this host issue (A-14); the
  `Dev Container build` CI job builds it before merge.
- The first CI build found a real defect: the Go Feature has no `none` value for
  `golangciLintVersion` and failed looking up tag `vnone`. The Feature now installs the pinned
  2.14.0; `make` targets keep using `.tools/bin/golangci-lint`.

## Open questions

None.
