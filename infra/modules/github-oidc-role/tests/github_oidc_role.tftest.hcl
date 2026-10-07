mock_provider "aws" {}

variables {
  role_name          = "synthetic-test-role"
  oidc_provider_arn  = "arn:aws:iam::000000000000:oidc-provider/token.actions.githubusercontent.com"
  github_repository  = "example-org/example-repo"
  github_environment = "staging"
}

run "trust_policy_is_restricted" {
  command = plan

  assert {
    condition     = length(jsondecode(aws_iam_role.this.assume_role_policy).Statement) == 1
    error_message = "The trust policy must contain exactly one statement."
  }

  assert {
    condition     = jsondecode(aws_iam_role.this.assume_role_policy).Statement[0].Action == "sts:AssumeRoleWithWebIdentity"
    error_message = "Only web identity federation may be trusted."
  }

  assert {
    condition     = jsondecode(aws_iam_role.this.assume_role_policy).Statement[0].Condition.StringEquals["token.actions.githubusercontent.com:aud"] == "sts.amazonaws.com"
    error_message = "The audience must be sts.amazonaws.com."
  }

  assert {
    condition     = jsondecode(aws_iam_role.this.assume_role_policy).Statement[0].Condition.StringEquals["token.actions.githubusercontent.com:sub"] == "repo:example-org/example-repo:environment:staging"
    error_message = "The subject must be restricted to the repository and environment."
  }

  assert {
    condition     = !strcontains(aws_iam_role.this.assume_role_policy, "*") && !strcontains(aws_iam_role.this.assume_role_policy, "StringLike")
    error_message = "The trust policy must not use wildcards."
  }
}

run "rejects_wildcard_repository" {
  command = plan

  variables {
    github_repository = "example-org/*"
  }

  expect_failures = [var.github_repository]
}

run "rejects_unknown_environment" {
  command = plan

  variables {
    github_environment = "qa"
  }

  expect_failures = [var.github_environment]
}
