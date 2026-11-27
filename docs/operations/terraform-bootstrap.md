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

The F-07 API adds root variables without defaults. Set them as `TF_VAR_*` environment variables
in the same GitHub environment (A-32): `TF_VAR_portal_origin`, `TF_VAR_github_client_id` (the
OAuth App's public client ID, D-30; never the client secret), `TF_VAR_artifacts_bucket_name`,
and `TF_VAR_log_retention_in_days` (a Legal and DPO decision, D-10). `TF_VAR_api_package` is
set by the deployment workflow after it uploads the artifact. `TF_VAR_sign_in_enabled` stays
unset (`false`); production rejects `true` until D-20 is recorded.

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

## 7. First-deployment sequence

The `api` function references a versioned ZIP in the environment's artifact bucket, and that
bucket is created by the same root. The first deployment of an environment therefore runs in two
applies (A-32). The deployment workflow (F-03 step 5) implements exactly this sequence; nothing
below is run until F-03 step 5 is enabled, and production additionally requires D-10.

All commands run as the environment's deployment role, from the environment root
(`infra/environments/<environment>`), with the GitHub environment variables of section 5 set.
`<environment>` is `dev`, `staging`, or `production`.

1. **Phase A: storage, keys, tables, secret.** With `TF_VAR_api_package` unset, the plan contains
   the four KMS keys, the three tables, the GitHub client secret (without a value), and the
   artifact bucket, and no function or HTTP API.

   ```bash
   terraform init -input=false
   terraform plan -input=false -out=phase-a.tfplan
   terraform apply -input=false phase-a.tfplan
   ```

2. **Set the secret value.** The OAuth App owners (LKS organization admins, D-30) set the value of
   the secret named by `terraform output -raw github_client_secret_name` through the AWS console or
   `aws secretsmanager put-secret-value`. The value is never passed through Terraform, GitHub, or
   this repository. The resource policy denies reads to everyone except the function's role.

3. **Build and upload the versioned ZIP.** From the repository root, on the commit being deployed:

   ```bash
   make build-go
   (cd services/api/dist/api && zip -X -j ../api.zip bootstrap)
   aws s3api put-object \
     --bucket "$(terraform -chdir=infra/environments/<environment> output -raw artifacts_bucket_name)" \
     --key "api/$(git rev-parse HEAD).zip" \
     --body services/api/dist/api.zip \
     --server-side-encryption aws:kms \
     --query VersionId --output text
   ```

   The printed `VersionId` pins the artifact.

4. **Phase B: function and HTTP API.** With the key and version from step 3:

   ```bash
   export TF_VAR_api_package='{"s3_key":"api/<commit-sha>.zip","s3_object_version":"<VersionId>"}'
   terraform plan -input=false -out=phase-b.tfplan
   terraform apply -input=false phase-b.tfplan
   ```

   `terraform output -raw api_endpoint` is the execute-api endpoint that the CloudFront `/api/*`
   behaviour will use (D-02, D-16). Sign-in stays disabled (`TF_VAR_sign_in_enabled` unset).

Later deployments repeat steps 3 and 4 only. A redeployment that changes nothing but the artifact
replaces the function code in place; tables, keys, the secret, and the bucket are never recreated
(`prevent_destroy`, deletion protection).

## 8. Table-key authorization

The `core` and `audit` keys keep the default key policy, which delegates to IAM. Two identity
policies use them (A-32):

- The deployment role, as the table creator, needs `kms:Encrypt`, `kms:Decrypt`, `kms:ReEncrypt*`,
  `kms:GenerateDataKey*`, `kms:DescribeKey`, and `kms:CreateGrant` through
  `dynamodb.<region>.amazonaws.com` (F-03 step 5 adds them).
- The `api` function's role has `kms:Decrypt` and `kms:DescribeKey` on the two keys, only through
  `dynamodb.<region>.amazonaws.com`, only for this account, and only for the three table names in
  the encryption context (`api-service` data policy, tested). DynamoDB decrypts the table key on
  the function's behalf when its five-minute cache expires; the function never calls KMS directly.
