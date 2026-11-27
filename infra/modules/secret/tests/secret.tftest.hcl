mock_provider "aws" {}

variables {
  name             = "lamassu-cla/synthetic/github-oauth-client-secret"
  description      = "Synthetic test secret"
  kms_key_arn      = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"
  reader_role_arns = ["arn:aws:iam::000000000000:role/synthetic-api"]
}

run "no_value_and_customer_key" {
  command = plan

  assert {
    condition     = aws_secretsmanager_secret.this.kms_key_id == var.kms_key_arn
    error_message = "The secret must be encrypted with the customer-managed key."
  }

  assert {
    condition     = aws_secretsmanager_secret.this.recovery_window_in_days == 30
    error_message = "The default recovery window must be 30 days."
  }

  assert {
    condition     = length(jsondecode(output.resource_policy).Statement) == 1
    error_message = "The resource policy is deny-only; reads come from identity policies."
  }

  assert {
    condition     = length([for s in jsondecode(output.resource_policy).Statement : s if s.Sid == "DenyReadToOthers" && s.Effect == "Deny" && s.Action == "secretsmanager:GetSecretValue" && s.Condition.ArnNotEquals["aws:PrincipalArn"] == ["arn:aws:iam::000000000000:role/synthetic-api"]]) == 1
    error_message = "Every principal but the reader role must be denied GetSecretValue."
  }

  assert {
    condition     = alltrue([for s in jsondecode(output.resource_policy).Statement : s.Action == "secretsmanager:GetSecretValue"])
    error_message = "The resource policy governs reads only; management stays with IAM."
  }
}

run "rejects_immediate_deletion" {
  command = plan

  variables {
    recovery_window_in_days = 0
  }

  expect_failures = [var.recovery_window_in_days]
}

run "rejects_non_role_readers" {
  command = plan

  variables {
    reader_role_arns = ["arn:aws:iam::000000000000:user/someone"]
  }

  expect_failures = [var.reader_role_arns]
}
