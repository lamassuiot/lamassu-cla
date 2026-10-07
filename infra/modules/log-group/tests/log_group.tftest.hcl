mock_provider "aws" {}

variables {
  name              = "/synthetic/test"
  retention_in_days = 7
  kms_key_arn       = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"
}

run "encrypted_with_retention" {
  command = plan

  assert {
    condition     = aws_cloudwatch_log_group.this.kms_key_id == var.kms_key_arn
    error_message = "The log group must be encrypted with the customer-managed key."
  }

  assert {
    condition     = aws_cloudwatch_log_group.this.retention_in_days == 7
    error_message = "The configured retention must be applied."
  }
}

run "rejects_never_expire" {
  command = plan

  variables {
    retention_in_days = 0
  }

  expect_failures = [var.retention_in_days]
}
