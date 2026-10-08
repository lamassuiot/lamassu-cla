# Deliberately insecure fixture for scripts/check-terraform-scan.test.sh. Never deploy.
terraform {
  required_version = ">= 1.11"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.67.0, < 7.0.0"
    }
  }
}

resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
}

# Trust without audience or subject conditions. Checkov 3.3.26 does not report this when the
# provider ARN is a reference; the github-oidc-role plan tests cover it.
resource "aws_iam_role" "deploy" {
  name = "synthetic-insecure-fixture"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRoleWithWebIdentity"
      Principal = { Federated = aws_iam_openid_connect_provider.github.arn }
    }]
  })
}

resource "aws_iam_role_policy" "admin" {
  name = "synthetic-insecure-fixture"
  role = aws_iam_role.deploy.id
  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [{ Effect = "Allow", Action = "*", Resource = "*" }]
  })
}
