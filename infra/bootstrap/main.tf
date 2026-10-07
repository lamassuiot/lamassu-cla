provider "aws" {
  region = var.region

  default_tags {
    tags = {
      project      = "lamassu-cla"
      environment  = var.environment
      "managed-by" = "terraform"
    }
  }
}

locals {
  github_oidc_url = "https://token.actions.githubusercontent.com"
  oidc_provider_arn = (
    var.create_github_oidc_provider
    ? aws_iam_openid_connect_provider.github[0].arn
    : data.aws_iam_openid_connect_provider.github[0].arn
  )
}

resource "aws_iam_openid_connect_provider" "github" {
  count = var.create_github_oidc_provider ? 1 : 0

  url            = local.github_oidc_url
  client_id_list = ["sts.amazonaws.com"]
}

data "aws_iam_openid_connect_provider" "github" {
  count = var.create_github_oidc_provider ? 0 : 1

  url = local.github_oidc_url
}

module "state_key" {
  source = "../modules/kms-key"

  description = "Terraform state encryption (${var.environment})"
  alias_name  = "alias/lamassu-cla-${var.environment}-terraform-state"
}

module "state_bucket" {
  source = "../modules/s3-bucket"

  bucket_name                        = var.state_bucket_name
  kms_key_arn                        = module.state_key.key_arn
  noncurrent_version_expiration_days = var.state_noncurrent_version_expiration_days
}

module "deployment_role" {
  source = "../modules/github-oidc-role"

  role_name          = "lamassu-cla-${var.environment}-deploy"
  oidc_provider_arn  = local.oidc_provider_arn
  github_repository  = var.github_repository
  github_environment = var.environment

  # Phase 2 environment roots create no resources, so the role only needs its own state.
  # Each feature that adds resources extends this policy with the permissions it needs.
  inline_policies = {
    terraform-state = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Sid      = "ListEnvironmentState"
          Effect   = "Allow"
          Action   = "s3:ListBucket"
          Resource = module.state_bucket.bucket_arn
          Condition = {
            StringLike = { "s3:prefix" = ["${var.environment_state_prefix}*"] }
          }
        },
        {
          Sid      = "ReadWriteEnvironmentState"
          Effect   = "Allow"
          Action   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
          Resource = "${module.state_bucket.bucket_arn}/${var.environment_state_prefix}*"
        },
        {
          Sid      = "UseStateKey"
          Effect   = "Allow"
          Action   = ["kms:Decrypt", "kms:Encrypt", "kms:GenerateDataKey"]
          Resource = module.state_key.key_arn
        },
      ]
    })
  }
}
