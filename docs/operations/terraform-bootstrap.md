# Terraform Bootstrap and Deployment Prerequisites

This guide lists every manual step the platform owner performs before any environment is
deployed ([F-03](../../specs/features/F-03-terraform-foundation.md)). Nothing in this guide is
automated, and no workflow in this repository creates AWS resources.

Do not commit account IDs, ARNs, bucket names, role names, or any other environment identifier.
Record them only in GitHub environment variables and in the platform owner's private records.

> **Do not run this procedure yet.** It describes the merged per-account bootstrap. The proposed
> central state account (D-24) changes it (contradiction C-01 in
> [`STATE.md`](../../specs/STATE.md#contradictions)). No bootstrap, plan against AWS, or apply may
> run until the platform owner confirms the Infrastructure/Tooling account, the workload accounts,
> the Control Tower controls, and the GitHub deployment environments.

## 1. Prerequisites per environment

Complete and record each item for `dev`, `staging`, and `production` (D-04).

| # | Prerequisite | Owner |
| --- | --- | --- |
| 1 | A dedicated AWS account exists for the environment, created outside this repository. | Platform owner |
| 2 | The region is `eu-west-1`, or `eu-south-2` if required. No other region is permitted. | Platform owner |
| 3 | CloudTrail is verified (section 2). An AWS Organizations baseline is not assumed. | Platform owner |
| 4 | Account-level S3 Block Public Access is verified (section 3). An AWS Organizations baseline is not assumed. | Platform owner |
| 5 | `infra/bootstrap` has been applied in the account (section 4). | Platform owner |
| 6 | The GitHub environment exists with its variables (section 5). | Repository administrator |

Production additionally requires Legal confirmation of the Object Lock mode and retention (D-10).
The production root refuses to plan until those values are supplied.

## 2. Verify CloudTrail

Run with administrator credentials for the account and region:

```bash
aws cloudtrail describe-trails --include-shadow-trails
aws cloudtrail get-trail-status --name <trail-name-or-arn>
```

The account passes when at least one trail is multi-Region (`IsMultiRegionTrail: true`), logs
management events, and reports `IsLogging: true`. If no trail qualifies, the platform owner creates
one before continuing. Record the result, not the trail ARN, in the environment checklist.

## 3. Verify account-level S3 Block Public Access

```bash
aws s3control get-public-access-block --account-id <account-id>
```

The account passes when all four settings (`BlockPublicAcls`, `IgnorePublicAcls`,
`BlockPublicPolicy`, `RestrictPublicBuckets`) are `true`. Every bucket created by this repository
also sets bucket-level Block Public Access explicitly (AC-03-9); the account setting is an
additional layer, not a replacement.

## 4. Apply `infra/bootstrap`

The bootstrap root creates the Terraform state bucket, its KMS key, and the GitHub OIDC deployment
role for one environment. Run it once per account, from a clean clone of `main`, with administrator
credentials for that account.

### 4.1 Choose the inputs

| Variable | Value |
| --- | --- |
| `environment` | `dev`, `staging`, or `production`. |
| `region` | `eu-west-1` (default) or `eu-south-2`. |
| `state_bucket_name` | A globally unique name chosen by the platform owner. |
| `state_noncurrent_version_expiration_days` | How long superseded state versions are kept. Chosen by the platform owner. |
| `create_github_oidc_provider` | `false` only if the account already has the GitHub Actions OIDC provider. |

Put the values in a file outside the repository, for example `~/lamassu-cla-bootstrap/dev.tfvars`.

### 4.2 First apply with local state

The S3 backend cannot be used before its bucket exists, so the first apply uses local state. Create
a temporary override file; `*_override.tf` files are ignored by Git.

```bash
cd infra/bootstrap
printf 'terraform {\n  backend "local" {}\n}\n' > backend_override.tf
terraform init
terraform plan -var-file="$HOME/lamassu-cla-bootstrap/dev.tfvars" -out=bootstrap.tfplan
terraform apply bootstrap.tfplan
terraform output
```

Review the plan before applying. It must create only the OIDC provider (unless disabled), one KMS
key and alias, one bucket with its controls, and one IAM role with one inline policy.

### 4.3 Migrate the bootstrap state into the bucket

```bash
rm backend_override.tf
terraform init -migrate-state \
  -backend-config="bucket=<state_bucket_name>" \
  -backend-config="key=bootstrap/terraform.tfstate" \
  -backend-config="region=<region>"
terraform plan -var-file="$HOME/lamassu-cla-bootstrap/dev.tfvars"
```

Confirm the migration when prompted. The final plan must report no changes. Then delete the local
`terraform.tfstate` and `terraform.tfstate.backup` files and `bootstrap.tfplan`.

The `bootstrap/` prefix is reserved for this state. The deployment role can only access objects
under `environment/` (the `environment_state_prefix` variable), so deployments cannot change the
bootstrap resources.

### 4.4 Protection

The state bucket and KMS key use `prevent_destroy`, the bucket has `force_destroy` disabled, and
the key has a 30-day deletion window. Never remove these protections to destroy bootstrap
resources.

## 5. Configure the GitHub environment

Create the GitHub environment named exactly `dev`, `staging`, or `production`. The OIDC role trusts
only `repo:lamassuiot/lamassu-cla:environment:<environment>`.

Set these environment variables from `terraform output`:

| Variable | Source |
| --- | --- |
| `AWS_REGION` | `region` |
| `AWS_ROLE_ARN` | `deployment_role_arn` |
| `TF_STATE_BUCKET` | `state_bucket_name` |
| `TF_STATE_KEY` | `environment_state_key` |

For `production`, configure required reviewers on the environment (AC-03-10).

The deployment workflow (F-03 implementation plan, step 5) is not part of this repository yet. It is
added and enabled only after every prerequisite above is complete; production stays blocked until
D-10 is confirmed.

## 6. Local validation without AWS

Contributors validate Terraform without credentials or AWS access:

```bash
make check-terraform
```

This runs `terraform fmt`, `validate` with `-backend=false`, plan-only `terraform test` with mocked
providers, TFLint, and the Checkov configuration scan with its insecure-fixture test. The Dev
Container provides Terraform, TFLint, and Python 3.12.
