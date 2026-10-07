mock_provider "aws" {}

variables {
  description = "Synthetic test key"
  alias_name  = "alias/synthetic-test-key"
}

run "secure_defaults" {
  command = plan

  assert {
    condition     = aws_kms_key.this.enable_key_rotation
    error_message = "Automatic key rotation must be enabled."
  }

  assert {
    condition     = aws_kms_key.this.deletion_window_in_days == 30
    error_message = "The default deletion window must be 30 days."
  }

  assert {
    condition     = aws_kms_alias.this.name == "alias/synthetic-test-key"
    error_message = "The alias must use the configured name."
  }
}

run "rejects_short_deletion_window" {
  command = plan

  variables {
    deletion_window_in_days = 6
  }

  expect_failures = [var.deletion_window_in_days]
}

run "rejects_reserved_alias" {
  command = plan

  variables {
    alias_name = "alias/aws/s3"
  }

  expect_failures = [var.alias_name]
}
