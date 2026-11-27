mock_provider "aws" {
  # Synthetic account: the api-service composition derives ARNs from it.
  mock_data "aws_caller_identity" {
    defaults = { account_id = "000000000000" }
  }

  mock_data "aws_region" {
    defaults = { region = "eu-west-1", name = "eu-west-1" }
  }
}

# Synthetic values for the F-07 variables that have no defaults; real values come from GitHub
# environment variables at plan time (D-05).
variables {
  portal_origin         = "https://cla.example"
  github_client_id      = "synthetic-client-id"
  artifacts_bucket_name = "synthetic-artifacts"
  log_retention_in_days = 30

  api_package = {
    s3_key            = "api/0123456789abcdef.zip"
    s3_object_version = "synthetic-version"
  }
}

run "fails_without_object_lock_values" {
  command = plan

  expect_failures = [var.evidence_object_lock]
}

run "accepts_legal_confirmed_values" {
  command = plan

  # Synthetic values for the test only; production values come from Legal (D-10).
  variables {
    evidence_object_lock = {
      mode = "COMPLIANCE"
      days = 1
    }
  }

  assert {
    condition     = output.environment == "production" && output.region == "eu-west-1"
    error_message = "The production root must default to eu-west-1."
  }
}

run "rejects_unknown_region" {
  command = plan

  variables {
    region = "eu-central-1"
    evidence_object_lock = {
      mode = "COMPLIANCE"
      days = 1
    }
  }

  expect_failures = [var.region]
}

run "rejects_sign_in_before_d20" {
  command = plan

  variables {
    sign_in_enabled = true
    evidence_object_lock = {
      mode = "COMPLIANCE"
      days = 1
    }
  }

  expect_failures = [var.sign_in_enabled]
}

run "phase_a_without_artifact" {
  command = plan

  variables {
    api_package = null
    evidence_object_lock = {
      mode = "COMPLIANCE"
      days = 1
    }
  }

  assert {
    condition     = output.api_endpoint == null && output.github_client_secret_name == "lamassu-cla/production/github-oauth-client-secret"
    error_message = "Without an artifact the root plans storage, keys, tables, and the secret only (A-32)."
  }
}
