# Local Development Setup

This guide covers the tooling foundation (F-01). Application features are added from Phase 3.
See [`specs/features/F-01-monorepo-tooling.md`](../../specs/features/F-01-monorepo-tooling.md).

## Prerequisites

| Tool | Version | Notes |
| --- | --- | --- |
| Node.js | 24 (from `.nvmrc`) | Install with [nvm](https://github.com/nvm-sh/nvm) and run `nvm use`. Other major versions are rejected (`engine-strict`). |
| npm | 11 or later | Ships with Node.js 24. |
| Go | Any Go 1.21 or later | The `Makefile` sets `GOTOOLCHAIN=go1.27.1`; Go downloads that toolchain automatically. |
| GNU Make | Any recent version | |
| Docker | Any recent version | Only for the Dev Container and local emulators. |

Terraform 1.16, TFLint, and Python 3.12 (for Checkov) are needed only for infrastructure work
(F-03); the Dev Container includes them.

## First run

```bash
nvm use
make bootstrap
make check
```

`make bootstrap` installs npm dependencies from `package-lock.json` (`npm ci`) and installs the
pinned `golangci-lint` and `govulncheck` into `.tools/bin`. `make check` runs lint, tests, and
builds as CI does.

## Make targets

| Target | What it does |
| --- | --- |
| `make help` | Lists targets. |
| `make bootstrap` | Installs npm dependencies and pinned Go tools. |
| `make tools` | Installs pinned Go tools into `.tools/bin` only. |
| `make lint` | Biome lint and format check (`biome ci`), TypeScript type check, the design-token drift check, Redocly lint of the OpenAPI contract, `go vet`, `golangci-lint` (including hexagonal-layering rules), the private-URL scan, and the TypeScript directive check. |
| `make lint-openapi` | Lints `specs/api/openapi.yaml` with Redocly using `redocly.yaml`. |
| `make test` | Vitest in every workspace, `go test -race`, and the script tests (private-URL scan, TypeScript directives, and layering rules). |
| `make vuln-go` | Runs the pinned `govulncheck` against the Go module and toolchain. Needs network access to the Go vulnerability database; not part of `make check`. |
| `make build` | Builds the portal with Vite and the Lambda binaries (`linux/arm64`) into `services/api/dist/`. |
| `make format` | Formats web sources and applies safe Biome fixes (`biome check --write`), and formats Go sources. |
| `make check` | `lint`, `test`, and `build`. |
| `make clean` | Removes build output. |
| `make check-terraform` | `fmt-terraform`, `validate-terraform`, `test-terraform`, `lint-terraform`, and `scan-terraform`. Needs Terraform and TFLint (Dev Container) but no AWS access; not part of `make check`. |
| `make test-terraform` | Validates every module and root without a backend and runs plan-only `terraform test` with mocked providers. |
| `make lint-terraform` | TFLint with the pinned AWS ruleset (`.tflint.hcl`). |
| `make scan-terraform` | Runs `test-terraform-scan`, then scans `infra/` with Checkov. Any failed check fails the scan; exceptions are inline `checkov:skip=<ID>:<reason>` comments. Installs Checkov into `.tools/checkov` from the hash-locked `scripts/checkov/requirements.txt`, and runs it with an empty environment (no AWS credentials). CI also runs the scan without network access. |
| `make test-terraform-scan` | Checks that Checkov reports the expected findings in the insecure fixtures under `scripts/testdata/checkov/`, and that every inline skip has a reason. |
| `make lock-checkov` | Regenerates `scripts/checkov/requirements.txt` from `requirements.in` with pip-tools. Needs Python 3.12 and network access. |

The [Terraform bootstrap guide](../operations/terraform-bootstrap.md) covers the platform-owner
steps that precede any deployment.

## Workspaces

| Path | Contents |
| --- | --- |
| `apps/web` | Portal (Vite, React, TypeScript). Run `npm run dev -w apps/web` and open <http://localhost:5173>. |
| `packages/design-system` | Design tokens (F-04) and components (F-05). Tokens are defined in `src/tokens/`; after changing them, run `npm run tokens -w @lamassu-cla/design-system` and commit the regenerated `src/styles/tokens.css`. `make lint` fails while the file is out of date. |
| `packages/api-client` | Typed API client generated from OpenAPI (F-02; generation deferred by A-18). |
| `packages/shared-types` | Types shared across workspaces. |
| `services/api` | Go module for the Lambda functions. Run `go test ./...` from this directory. |

## Hexagonal layering rules

`golangci-lint` uses `depguard` to enforce the layering in
[`specs/architecture.md`](../../specs/architecture.md#31-layers):

| Package | May import |
| --- | --- |
| `internal/domain` | Standard library (without I/O packages) and `domain` only. |
| `internal/ports` | Standard library and `domain`. |
| `internal/application` | Standard library, `domain`, and `ports`. |
| `internal/handlers`, `internal/middleware` | Anything except `adapters`. |
| `internal/adapters/...` | Anything, including SDKs. |
| Anything outside `adapters` and `cmd` | No AWS, GitHub, or signing-provider SDKs. |

`scripts/check-go-layering.test.sh` proves these rules reject violations.

## Dev Container

The `.devcontainer/` configuration provides Node.js 24, Go 1.27, Terraform 1.16, and TFLint on
a pinned Ubuntu 24.04 base image, and runs `make bootstrap` after creation. Open the repository in
VS Code and choose **Reopen in Container**.

The Terraform feature declares a dependency on the GitHub CLI feature, so the container also
includes `gh`. Feature versions and digests, including that dependency, are pinned in
`.devcontainer/devcontainer-lock.json`.

The container build downloads packages, so Docker must be able to resolve external hostnames. If
`apt` reports `Temporary failure resolving`, configure DNS for the Docker daemon on the host.

The `Dev Container build` CI job builds the container from the lockfile and runs `make check` in
it on every pull request.

## Rules for local work

- Never commit `.env` files, credentials, AWS account identifiers, or real personal data. Use
  synthetic data in tests and fixtures.
- Dependencies are pinned to exact versions (`save-exact`). Commit `package-lock.json` and
  `go.sum` changes together with the manifest change.
