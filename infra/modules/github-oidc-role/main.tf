locals {
  token_host = "token.actions.githubusercontent.com"

  trust_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "GitHubActionsEnvironment"
        Effect    = "Allow"
        Principal = { Federated = var.oidc_provider_arn }
        Action    = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${local.token_host}:aud" = "sts.amazonaws.com"
            "${local.token_host}:sub" = "repo:${var.github_repository}:environment:${var.github_environment}"
          }
        }
      },
    ]
  })
}

resource "aws_iam_role" "this" {
  name                 = var.role_name
  assume_role_policy   = local.trust_policy
  max_session_duration = var.max_session_duration
  tags                 = var.tags
}

resource "aws_iam_role_policy" "this" {
  for_each = var.inline_policies

  name   = each.key
  role   = aws_iam_role.this.id
  policy = each.value
}
