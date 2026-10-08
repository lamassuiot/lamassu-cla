# ADR-0012: Checkov for the Terraform security scan

- Status: Proposed
- Date: 2026-10-08
- Decision-log entry: A-19 (replaces Trivy in the proposed A-19 tool pins)

## Context

The repository owner approved A-19 on 2026-10-08 on the condition that Checkov replaces Trivy as
the Terraform configuration scanner. The conditions are an evaluated, pinned, and
integrity-verified installation, no AWS credentials, failure on high and critical findings,
`terraform validate`, TFLint, and plan tests kept as independent controls, and proof that the scan
detects deliberately vulnerable Terraform before A-19 is updated.

The evaluation used Checkov 3.3.26, the latest release on PyPI (2026-10-07, Apache-2.0), with
Python 3.12. Checkov 3.3.26 declares support for Python 3.9 to 3.12. The evaluation found:

- **Installation.** `checkov==3.3.26` resolves to 96 packages. A `pip-compile --generate-hashes`
  lock installs with `--require-hashes --only-binary=:all: --no-deps`, and `pip check` passes.
- **Severities.** Without a Prisma Cloud API key, every finding has no severity. With
  `--hard-fail-on HIGH,CRITICAL`, a fixture with an `"*"` administrator policy and disabled
  Block Public Access exited 0. Without that option, the same fixture exited 1 with 26 failed
  checks.
- **Offline.** Scans run inside a container without network access and without AWS variables.
- **Dependency vulnerabilities.** `pip-audit` 2.10.0 reported three advisories:

  | Advisory | Package | Severity | Fix | Use in Checkov |
  | --- | --- | --- | --- | --- |
  | GHSA-9w56-46f6-3qhx | `asteval` 1.0.6 | Medium | 1.0.9 | Evaluates Terraform expressions from scanned files (`safe_eval_functions.py`). |
  | GHSA-89v8-rhwq-hf77 (CVE-2026-55244) | `asteval` 1.0.6 | Medium | 1.0.9 | Same. |
  | GHSA-wj6h-64fc-37mp (CVE-2024-23342) | `ecdsa` 0.19.2 | High | None | Verifies external-check signatures only; verification is unaffected. |

  Checkov 3.3.26 and its development branch pin `asteval==1.0.6` exactly, so no Checkov release
  avoids the `asteval` advisories.
- **Coverage gap.** Checkov 3.3.26 did not report a GitHub OIDC trust policy without audience or
  subject conditions when the provider ARN is a reference. The `github-oidc-role` plan tests cover
  that case.
- **Existing code.** `infra/` failed four checks on the bootstrap state bucket: CKV_AWS_300,
  CKV_AWS_144, CKV_AWS_18, and CKV2_AWS_62.

## Decision

- Checkov 3.3.26 replaces Trivy. `scripts/checkov/requirements.in` pins it, and
  `scripts/checkov/requirements.txt` locks all 96 packages with SHA-256 hashes for Python 3.12.
  `make scan-terraform` installs only from that lock, with `--require-hashes`,
  `--only-binary=:all:`, and `--no-deps`, into `.tools/checkov`, then runs `pip check`.
- The scan runs with an empty environment (`env -i`), so no AWS or Prisma Cloud credentials reach
  it. It uses `--skip-download` and `--download-external-modules false`. The CI job has no
  secrets and no `id-token` permission.
- **Failure policy.** Severity gating is not possible offline, so every failed check fails the
  scan. This is stricter than failing on high and critical findings. `--soft-fail` and
  `--hard-fail-on` are not used. Exceptions are inline `checkov:skip=<ID>:<reason>` comments,
  placed on the module call so that the reusable module stays strict. A skip without a reason
  fails the fixture test.
- **Fixture test.** `scripts/testdata/checkov/` holds deliberately insecure Terraform that is never
  deployed. `make test-terraform-scan` runs first and fails unless Checkov exits non-zero and
  reports the expected check IDs for public bucket access, missing KMS encryption and versioning,
  KMS rotation, log-group encryption and retention, and administrator IAM policies.
- **Existing findings.** The S3 module aborts incomplete multipart uploads after 7 days
  (CKV_AWS_300, input `abort_incomplete_multipart_upload_days`). CKV_AWS_144, CKV_AWS_18, and
  CKV2_AWS_62 are skipped on the bootstrap `state_bucket` call, with D-24 as the reason.
- `terraform validate`, plan-only `terraform test`, and TFLint remain separate Make targets and CI
  jobs. The `Terraform security scan` job name is unchanged.
- Python 3.12 comes from the Dev Container Python feature 1.8.0 (`os-provided`, Ubuntu 24.04) and,
  in CI, from `actions/setup-python` pinned by commit SHA. Dependabot updates the lock (`pip`
  ecosystem, `/scripts/checkov`). `make lock-checkov` regenerates it with pip-tools 7.5.1.

### Temporary dependency-review exception

The required `Dependency review` check fails on these advisories. It allows exactly the three
GHSA identifiers above (`allow-ghsas` in `.github/workflows/ci.yml`). No other advisory is
allowed, and the severity threshold is unchanged.

Rationale: the scan reads only Terraform from this repository, which is reviewed, and runs in a
credential-free, isolated job whose token is read-only. Any pull request can already run code in
that job through the `Makefile`, so the `asteval` sandbox escape adds no new capability there. The
`ecdsa` advisory concerns signing, which Checkov does not use.

- **Owner:** Maintainers track upstream fixes. The Security owner reviews the exception.
- **Review:** when a Checkov release requires `asteval>=1.0.9`, or at the start of Phase 3,
  whichever comes first. At review, remove the `asteval` identifiers once the lock is free of
  them; keep `ecdsa` only while its advisory has no fix and Checkov still uses it only for
  verification.
- Dependabot alerts for these advisories are dismissed as tolerable risk with a reference to this
  ADR. This is a manual repository setting.

## Consequences

- The scan fails on findings that a severity threshold would ignore. Each exception needs a
  justified inline skip, which is visible in review.
- Contributors without the Dev Container need Python 3.12 with `venv` to run `make scan-terraform`.
- The first scan downloads about 96 wheels from PyPI. Later runs reuse `.tools/checkov` until the
  lock changes.
- The dependency-review exception is repository-wide for the three identifiers. If another
  manifest adds a package with the same advisory, the exception also allows it.
- A-19 and this ADR stay proposed until the Security owner and the platform owner review the final
  wording.

## Alternatives considered

- **Keep Trivy.** Rejected by the repository owner.
- **Severity gating through Prisma Cloud.** Rejected: it needs an external account and a CI
  secret, and it sends scan metadata to a third party.
- **Official Docker image pinned by digest.** Rejected by the repository owner. It contains the
  same packages, which dependency review cannot see, and it needs Docker inside the Dev Container.
- **Defer Checkov until `asteval` is fixed.** Rejected by the repository owner in favor of this
  time-limited exception.
