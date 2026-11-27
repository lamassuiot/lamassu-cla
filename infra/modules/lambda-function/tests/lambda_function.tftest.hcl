mock_provider "aws" {}

# The log group ARN feeds the logging policy; a known value lets plan-only runs inspect it.
override_module {
  target = module.log_group
  outputs = {
    name = "/aws/lambda/synthetic-api"
    arn  = "arn:aws:logs:eu-west-1:000000000000:log-group:/aws/lambda/synthetic-api"
  }
}

variables {
  function_name                  = "synthetic-api"
  description                    = "Synthetic test function"
  reserved_concurrent_executions = 10
  log_retention_in_days          = 30
  kms_key_arn                    = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"

  package = {
    s3_bucket         = "synthetic-artifacts"
    s3_key            = "api/0123456789abcdef.zip"
    s3_object_version = "synthetic-version"
  }

  environment_variables = {
    CLA_SIGN_IN_ENABLED = "false"
  }
}

run "go_runtime_and_secure_defaults" {
  command = plan

  assert {
    condition     = aws_lambda_function.this.runtime == "provided.al2023" && aws_lambda_function.this.architectures == tolist(["arm64"]) && aws_lambda_function.this.handler == "bootstrap"
    error_message = "Functions run the Go custom runtime on arm64 (architecture 2)."
  }

  assert {
    condition     = aws_lambda_function.this.s3_bucket == "synthetic-artifacts" && aws_lambda_function.this.s3_key == "api/0123456789abcdef.zip" && aws_lambda_function.this.s3_object_version == "synthetic-version"
    error_message = "The artifact is the pinned S3 object version."
  }

  assert {
    condition     = aws_lambda_function.this.reserved_concurrent_executions == 10
    error_message = "Reserved concurrency must be set."
  }

  assert {
    condition     = aws_lambda_function.this.kms_key_arn == var.kms_key_arn && one(aws_lambda_function.this.environment).variables["CLA_SIGN_IN_ENABLED"] == "false"
    error_message = "Environment variables are encrypted with the customer-managed key."
  }

  assert {
    condition     = one(aws_lambda_function.this.logging_config).log_format == "JSON" && one(aws_lambda_function.this.logging_config).log_group == "/aws/lambda/synthetic-api"
    error_message = "Structured logs go to the function's own log group."
  }

  assert {
    condition     = aws_lambda_function.this.memory_size == 512 && aws_lambda_function.this.timeout == 10
    error_message = "Default memory and timeout."
  }

  assert {
    condition     = output.log_group_name == "/aws/lambda/synthetic-api"
    error_message = "The log group is the function's own."
  }
}

run "least_privilege_role" {
  command = plan

  variables {
    inline_policies = {
      tables = jsonencode({ Version = "2012-10-17", Statement = [] })
    }
  }

  assert {
    condition     = jsondecode(aws_iam_role.this.assume_role_policy).Statement[0].Principal.Service == "lambda.amazonaws.com"
    error_message = "Only the Lambda service assumes the role."
  }

  assert {
    condition     = [for s in jsondecode(output.logging_policy).Statement : s.Sid] == ["WriteOwnLogs", "EncryptEnvironment"]
    error_message = "The logging policy grants log writes and key decryption only."
  }

  assert {
    condition     = alltrue([for s in jsondecode(output.logging_policy).Statement : s.Resource != "*"])
    error_message = "No statement in the logging policy uses a wildcard resource."
  }

  assert {
    condition     = jsondecode(output.logging_policy).Statement[0].Resource == "arn:aws:logs:eu-west-1:000000000000:log-group:/aws/lambda/synthetic-api:*"
    error_message = "Log writes are scoped to the function's own log group."
  }

  assert {
    condition     = keys(aws_iam_role_policy.this) == ["tables"] && aws_iam_role_policy.logging.name == "logging"
    error_message = "Inline policies are attached by name next to the logging policy."
  }
}

run "rejects_unlimited_concurrency" {
  command = plan

  variables {
    reserved_concurrent_executions = -1
  }

  expect_failures = [var.reserved_concurrent_executions]
}

run "rejects_non_zip_package" {
  command = plan

  variables {
    package = {
      s3_bucket         = "synthetic-artifacts"
      s3_key            = "api/bootstrap"
      s3_object_version = "synthetic-version"
    }
  }

  expect_failures = [var.package]
}
