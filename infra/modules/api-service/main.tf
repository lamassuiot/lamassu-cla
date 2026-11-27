# Composition of the F-07 sign-in API: keys per data class, the three DynamoDB tables (ADR-0009),
# the GitHub OAuth client secret, the `api` function, and the HTTP API in front of it. Plan-only
# until F-03 step 5; nothing here authorizes deployment or real sign-ins (D-20).

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  prefix     = "lamassu-cla-${var.environment}"
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.region

  api_function_name = "${local.prefix}-api"
  secret_name       = "lamassu-cla/${var.environment}/github-oauth-client-secret"

  # CloudWatch Logs encrypts with the operations key only for this environment's log groups.
  operations_key_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AccountAdministration"
        Effect    = "Allow"
        Principal = { AWS = "arn:aws:iam::${local.account_id}:root" }
        Action    = "kms:*"
        Resource  = "*"
      },
      {
        Sid       = "CloudWatchLogs"
        Effect    = "Allow"
        Principal = { Service = "logs.${local.region}.amazonaws.com" }
        Action    = ["kms:Encrypt", "kms:Decrypt", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:DescribeKey"]
        Resource  = "*"
        Condition = {
          ArnLike = {
            "kms:EncryptionContext:aws:logs:arn" = "arn:aws:logs:${local.region}:${local.account_id}:log-group:/aws/*/${local.prefix}-*"
          }
        }
      },
    ]
  })

  api_environment = merge(
    {
      CLA_PORTAL_ORIGIN           = var.portal_origin
      CLA_GITHUB_CALLBACK_URL     = "${var.portal_origin}/api/auth/github/callback"
      CLA_SIGN_IN_ERROR_ROUTE     = var.sign_in_error_route
      CLA_SIGN_IN_ENABLED         = var.sign_in_enabled ? "true" : "false"
      CLA_RETURN_TO_ALLOWLIST     = join(",", var.return_to_allowlist)
      CLA_TABLE_EPHEMERAL         = module.ephemeral_table.name
      CLA_TABLE_CORE              = module.core_table.name
      CLA_TABLE_AUDIT             = module.audit_table.name
      CLA_GITHUB_CLIENT_ID        = var.github_client_id
      CLA_GITHUB_CLIENT_SECRET_ID = local.secret_name
    },
    var.session_idle_seconds == null ? {} : { CLA_SESSION_IDLE = tostring(var.session_idle_seconds) },
    var.session_absolute_seconds == null ? {} : { CLA_SESSION_ABSOLUTE = tostring(var.session_absolute_seconds) },
  )

  # Exactly what the F-07 use cases do (adapters/aws): consume and write attempts and sessions in
  # `ephemeral`, upsert the contributor profile in `core`, append audit events. The audit table is
  # append-only (ADR-0009): the deny binds even if a later policy grants more.
  api_data_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "EphemeralAttemptsAndSessions"
        Effect   = "Allow"
        Action   = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:UpdateItem", "dynamodb:DeleteItem", "dynamodb:ConditionCheckItem"]
        Resource = module.ephemeral_table.arn
      },
      {
        Sid      = "CoreContributorProfile"
        Effect   = "Allow"
        Action   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
        Resource = module.core_table.arn
      },
      {
        Sid      = "AuditAppend"
        Effect   = "Allow"
        Action   = "dynamodb:PutItem"
        Resource = module.audit_table.arn
      },
      {
        Sid      = "AuditAppendOnly"
        Effect   = "Deny"
        Action   = ["dynamodb:UpdateItem", "dynamodb:DeleteItem", "dynamodb:BatchWriteItem"]
        Resource = [module.audit_table.arn, "${module.audit_table.arn}/index/*"]
      },
      {
        Sid      = "ReadGitHubClientSecret"
        Effect   = "Allow"
        Action   = "secretsmanager:GetSecretValue"
        Resource = module.github_client_secret.arn
      },
      {
        Sid      = "DecryptSecrets"
        Effect   = "Allow"
        Action   = "kms:Decrypt"
        Resource = module.secrets_key.key_arn
        Condition = {
          StringEquals = { "kms:ViaService" = "secretsmanager.${local.region}.amazonaws.com" }
        }
      },
    ]
  })
}

# Keys per data class (security specification 4).

module "core_key" {
  source = "../kms-key"

  description = "DynamoDB core and ephemeral data (${var.environment})"
  alias_name  = "alias/${local.prefix}-core"
}

module "audit_key" {
  source = "../kms-key"

  description = "DynamoDB audit data (${var.environment})"
  alias_name  = "alias/${local.prefix}-audit"
}

module "secrets_key" {
  source = "../kms-key"

  description = "Secrets Manager secrets (${var.environment})"
  alias_name  = "alias/${local.prefix}-secrets"
}

module "operations_key" {
  source = "../kms-key"

  description = "Log groups, Lambda environment, and deployment artifacts (${var.environment})"
  alias_name  = "alias/${local.prefix}-operations"
  policy      = local.operations_key_policy
}

# Tables (ADR-0009, domain model 5).

module "core_table" {
  source = "../dynamodb-table"

  name        = "${local.prefix}-core"
  hash_key    = "PK"
  range_key   = "SK"
  kms_key_arn = module.core_key.key_arn
  attributes = {
    PK     = "S", SK = "S",
    GSI1PK = "S", GSI2PK = "S", GSI2SK = "S", GSI3PK = "S", GSI3SK = "S",
  }
  global_secondary_indexes = {
    GSI1 = { hash_key = "GSI1PK" }
    GSI2 = { hash_key = "GSI2PK", range_key = "GSI2SK" }
    GSI3 = { hash_key = "GSI3PK", range_key = "GSI3SK" }
  }
}

module "audit_table" {
  source = "../dynamodb-table"

  name             = "${local.prefix}-audit"
  hash_key         = "PK"
  range_key        = "SK"
  kms_key_arn      = module.audit_key.key_arn
  stream_view_type = "NEW_IMAGE"
  attributes = {
    PK     = "S", SK = "S",
    GSI1PK = "S", GSI1SK = "S", GSI2PK = "S", GSI2SK = "S",
  }
  global_secondary_indexes = {
    GSI1 = { hash_key = "GSI1PK", range_key = "GSI1SK" }
    GSI2 = { hash_key = "GSI2PK", range_key = "GSI2SK" }
  }
}

module "ephemeral_table" {
  source = "../dynamodb-table"

  name          = "${local.prefix}-ephemeral"
  hash_key      = "PK"
  kms_key_arn   = module.core_key.key_arn
  ttl_attribute = "expires_at"
  attributes    = { PK = "S" }
}

# GitHub OAuth client secret: value set by LKS organization admins (D-30), readable only by the
# `api` function's role.

module "github_client_secret" {
  source = "../secret"

  name             = local.secret_name
  description      = "GitHub OAuth App client secret for sign-in (${var.environment})"
  kms_key_arn      = module.secrets_key.key_arn
  reader_role_arns = [module.api.role_arn]
}

# Deployment artifacts, uploaded by the deployment workflow.

module "artifacts_bucket" {
  # checkov:skip=CKV_AWS_144:Build artifacts are reproducible from Git; no cross-region replication.
  # checkov:skip=CKV_AWS_18:Artifact bucket access logging is decided with the observability design (A-32).
  # checkov:skip=CKV2_AWS_62:No consumer of artifact event notifications exists.
  source = "../s3-bucket"

  bucket_name                        = var.artifacts_bucket_name
  kms_key_arn                        = module.operations_key.key_arn
  noncurrent_version_expiration_days = var.artifact_noncurrent_version_expiration_days
}

# The `api` function and its HTTP API.

module "api" {
  source = "../lambda-function"

  function_name                  = local.api_function_name
  description                    = "CLA platform API (${var.environment})"
  memory_size                    = var.api_memory_size
  timeout                        = var.api_timeout
  reserved_concurrent_executions = var.api_reserved_concurrency
  log_retention_in_days          = var.log_retention_in_days
  kms_key_arn                    = module.operations_key.key_arn
  environment_variables          = local.api_environment

  package = {
    s3_bucket         = module.artifacts_bucket.bucket_name
    s3_key            = var.api_package.s3_key
    s3_object_version = var.api_package.s3_object_version
  }

  inline_policies = {
    data = local.api_data_policy
  }
}

module "http_api" {
  source = "../http-api"

  name                   = local.api_function_name
  description            = "CLA platform HTTP API (${var.environment})"
  lambda_invoke_arn      = module.api.invoke_arn
  lambda_function_name   = module.api.function_name
  route_keys             = ["ANY /api/{proxy+}"]
  throttling_burst_limit = var.api_throttling_burst_limit
  throttling_rate_limit  = var.api_throttling_rate_limit
  log_retention_in_days  = var.log_retention_in_days
  kms_key_arn            = module.operations_key.key_arn
}
