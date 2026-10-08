# F-03: Terraform foundation and environments

| Field | Value |
| --- | --- |
| Status | Steps 1 to 4 merged in PR #20 (no deployment); Checkov replaces Trivy (PRs #22, #23; ADR-0012 proposed); step 5 blocked (see Preconditions); bootstrap rework pending D-24 (C-01) |
| Phase | 2 |
| Owner | Repository owner; platform owner for provisioning |
| Depends on | F-01; D-04, D-05, D-24; A-19; [ADR-0005](../../docs/decisions/0005-configurable-object-lock-retention.md), [ADR-0010](../../docs/decisions/0010-cloudfront-frontend-hosting.md), [ADR-0012](../../docs/decisions/0012-checkov-terraform-scan.md) |
| Approved by | Repository owner, 2026-10-05 (implementation approved; no deployment) |

## Problem statement

All infrastructure must be managed with Terraform, in separate AWS accounts per environment, with
remote state and keyless GitHub deployments, before application resources are added.

## User roles

Platform owner (provisions accounts and runs the bootstrap), maintainers (review plans),
deployment workflows (assume roles through OIDC).

## Preconditions

Platform-owner prerequisites. None may be assumed; each must be completed and confirmed before any
environment deployment:

1. One AWS account per environment (`dev`, `staging`, `production`) exists (D-04), created outside
   this repository by the platform owner.
2. The region is `eu-west-1`. `eu-south-2` is the only permitted alternative.
3. CloudTrail status is verified per account. An AWS Organizations baseline is **not** assumed.
4. Account-level S3 Block Public Access status is verified per account. An AWS Organizations
   baseline is **not** assumed.
5. The platform owner has run `infra/bootstrap` in each account.
6. A dedicated Infrastructure/Tooling account for Terraform state exists and is confirmed by the
   platform owner (D-24). Its existence is not assumed.
7. The Control Tower controls that apply to the tooling and workload accounts are confirmed by the
   platform owner.
8. The GitHub deployment environments (`dev`, `staging`, `production`) exist and are confirmed.

Production additionally requires Legal confirmation of the Object Lock configuration (D-10).

Implementing and validating the Terraform code (static checks and tests with mocked providers) does
not require these prerequisites. No branch bootstraps an account, plans against AWS, applies, or
creates AWS resources until they are met.

## Main flow

This flow describes the merged per-account bootstrap. It changes when D-24 is approved
(contradiction C-01 in [`STATE.md`](../STATE.md#contradictions)); do not run it before then.

1. The platform owner runs `infra/bootstrap` once per account with administrator credentials. It
   creates the state bucket and the GitHub OIDC deployment role for that environment.
2. The platform owner sets GitHub environment variables (role ARN, region, state bucket name) for
   `dev`, `staging`, and `production`. Nothing environment-specific is committed.
3. Deployment workflows assume the role through OIDC, run `terraform init` with partial backend
   configuration, then `plan` and `apply`.

## Alternative flows

- Bootstrap state: `infra/bootstrap` starts with local state, then migrates into the bucket it
  created (`terraform init -migrate-state`).

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Region not allowed | Variable validation | Plan fails | Allowed regions listed |
| Production Object Lock values missing | Variable without default plus validation | Plan fails | Message referencing D-10 |
| Concurrent apply | S3 lock file (`use_lockfile = true`) | Second run waits or fails | Lock error |
| OIDC token from another repository or environment | Role trust conditions | Assume-role denied | AWS access denied |

## Security considerations

- State buckets: versioning, SSE-KMS, S3 Block Public Access, TLS-only bucket policy, no public
  access, lifecycle expiry for old non-current versions only.
- Repository-level controls are enforced explicitly in Terraform for every resource, even when an
  account or organization baseline might already provide them: bucket-level Block Public Access,
  encryption, versioning, TLS-only policies, and least-privilege IAM.
- OIDC trust policies restrict the audience to `sts.amazonaws.com` and the subject to
  `repo:lamassuiot/lamassu-cla:environment:{environment}`.
- Deployment roles are scoped to the resources and services the platform uses. Any broad
  permission needed temporarily during Phase 2 is documented and removed before Phase 9.
- No account IDs, ARNs, bucket names, or domains are committed. Backend and provider settings come
  from partial configuration and variables.
- Production uses a GitHub environment with required reviewers.

## API changes

None.

## Data-model changes

None. DynamoDB tables are added by the features that use them.

## UI changes

None.

## Acceptance criteria

- **AC-03-1** Every root module declares `required_version >= 1.11` and an S3 backend with
  `use_lockfile = true` and no `dynamodb_table`.
- **AC-03-2** `infra/environments/{dev,staging,production}` contain composition and variables
  only; resources are defined in `infra/modules`.
- **AC-03-3** The `region` variable accepts only `eu-west-1` and `eu-south-2` and defaults to
  `eu-west-1`.
- **AC-03-4** All modules apply default tags (`project`, `environment`, `managed-by`) and secure
  defaults (encryption, Block Public Access, versioning).
- **AC-03-5** Object Lock variables have no default in production and fail validation when unset.
- **AC-03-6** `infra/bootstrap` creates the state bucket and the OIDC role, restricted to this
  repository and the named GitHub environment.
- **AC-03-7** `terraform fmt`, `validate`, TFLint, and the security scan pass for every root and
  module (F-02).
- **AC-03-8** `docs/operations/terraform-bootstrap.md` lists every manual step for the platform
  owner, including how to verify the CloudTrail and account-level S3 Block Public Access
  prerequisites.
- **AC-03-9** Every bucket module instance sets bucket-level Block Public Access, encryption,
  versioning, and a TLS-only policy explicitly, without relying on account-level settings.
- **AC-03-10** A production plan fails until the Legal-confirmed Object Lock mode and retention
  are provided (D-10). The production deployment job additionally requires the GitHub `production`
  environment approval.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F03-T1 | `terraform validate` per root module with `-backend=false`. | Static | AC-03-1, AC-03-2 |
| F03-T2 | `terraform test` with mocked providers: invalid region and missing production Object Lock values fail. | Module test | AC-03-3, AC-03-5 |
| F03-T3 | `terraform test`: the OIDC role trust policy contains the repository and environment conditions. | Module test | AC-03-6 |
| F03-T4 | Security scan reports no high findings. | Static | AC-03-4, AC-03-9 |
| F03-T5 | `terraform test`: the production root fails validation without Object Lock values. | Module test | AC-03-10 |
| F03-T6 | `terraform test`: the deployment role policy allows only state objects under `environment/` and the state key. | Module test | AC-03-6 |
| F03-T7 | `terraform test`: non-production roots reject compliance-mode Object Lock. | Module test | ADR-0005 |

## Observability requirements

CloudTrail is a platform-owner prerequisite verified per account (see Preconditions); it is not
assumed from an organization baseline. Application logging and alarms arrive with the resources
that need them.

## Rollback considerations

Bootstrap resources are protected with `prevent_destroy`. Environment roots in Phase 2 create no
data stores, so rollback is a revert and apply. State buckets are never deleted automatically.

## Implementation plan

1. **Implemented:** modules `kms-key`, `s3-bucket`, `github-oidc-role`, `log-group`, with tests.
2. **Implemented:** `infra/bootstrap`.
3. **Implemented:** environment roots with variables, validation, and default tags.
4. **Implemented:** [bootstrap and deployment guide](../../docs/operations/terraform-bootstrap.md).
5. Deployment workflow with OIDC and production approval. It is enabled only after the platform
   owner completes the prerequisites; production remains blocked until D-10 is confirmed.
   **Not started.**

Implementation notes (steps 1 to 4):

- Tests are plan-only (`command = plan`) with `mock_provider "aws"`; they need no credentials
  and never contact AWS. 28 runs cover the modules, the bootstrap root, and the three
  environment roots.
- No retention value is committed. Object Lock mode and days, log retention, and state
  non-current version expiry are required inputs supplied at deployment (ADR-0005, D-10).
  Non-production roots accept only governance mode; the production root fails without values.
- The deployment role created by `infra/bootstrap` may only read and write state objects under
  `environment/` and use the state KMS key. The `bootstrap/` state is out of its reach. Each
  feature that adds resources extends the policy with the permissions it needs (least privilege).
- The bootstrap root migrates from local to S3 state with a Git-ignored `backend_override.tf`.
- Environment roots compose no modules yet: Phase 2 creates no resources in them. The
  `evidence_object_lock` variable is validated now and consumed when the evidence bucket is added.
- Tool pins (A-19): AWS provider `>= 6.67.0, < 7.0.0` in every module, locked to 6.68.0 for
  `linux` and `darwin` on `amd64` and `arm64` in each root; TFLint AWS ruleset 0.49.0; Checkov
  3.3.26 from a hash-locked requirements file ([ADR-0012](../../docs/decisions/0012-checkov-terraform-scan.md)).
  Reusable modules do not commit lock files. Dependabot updates root lock files and the Checkov
  lock.
- Checkov fails on every failed check. The S3 module aborts incomplete multipart uploads after
  7 days (`abort_incomplete_multipart_upload_days`). The bootstrap state bucket skips CKV_AWS_144,
  CKV_AWS_18, and CKV2_AWS_62 inline, pending D-24.
- `make check-terraform` runs every check locally without AWS access. It is not part of
  `make check`, so contributors without Terraform are not blocked; CI runs it in the `Terraform`
  workflow (F-02 unit 5).

## Open questions

- D-24: central state account (platform owner, Security owner). Its approval resolves C-01 and
  defines the bootstrap rework.
- Deployment is blocked on the platform-owner prerequisites and, for production, on D-10.
