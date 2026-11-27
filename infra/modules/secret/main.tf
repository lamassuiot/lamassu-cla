# Secrets Manager secret without a value: the value is set outside Terraform by the people who
# own it (for the GitHub OAuth client secret, LKS organization admins; F-07, D-30), so it never
# enters state, pull requests, or CI logs.
resource "aws_secretsmanager_secret" "this" {
  # checkov:skip=CKV2_AWS_57:GitHub OAuth App client secrets cannot be rotated through an API; the owners rotate them by hand and update the secret value (F-07).
  name                    = var.name
  description             = var.description
  kms_key_id              = var.kms_key_arn
  recovery_window_in_days = var.recovery_window_in_days
  tags                    = var.tags
}

resource "aws_secretsmanager_secret_policy" "this" {
  secret_arn = aws_secretsmanager_secret.this.arn
  policy     = local.resource_policy
}

locals {
  # Only the listed reader roles can read the value; everyone else in the account can manage the
  # secret (set the value, rotate, delete) but not read it. The explicit deny also binds the
  # account root and administrators.
  resource_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowReaders"
        Effect    = "Allow"
        Principal = { AWS = var.reader_role_arns }
        Action    = "secretsmanager:GetSecretValue"
        Resource  = "*"
      },
      {
        Sid       = "DenyReadToOthers"
        Effect    = "Deny"
        Principal = { AWS = "*" }
        Action    = "secretsmanager:GetSecretValue"
        Resource  = "*"
        Condition = {
          ArnNotEquals = { "aws:PrincipalArn" = var.reader_role_arns }
        }
      },
    ]
  })
}
