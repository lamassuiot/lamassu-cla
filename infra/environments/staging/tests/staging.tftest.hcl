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

run "plans_without_object_lock_values" {
  command = plan

  assert {
    condition     = output.environment == "staging" && output.region == "eu-west-1"
    error_message = "The staging root must default to eu-west-1."
  }
}

run "accepts_eu_south_2" {
  command = plan

  variables {
    region = "eu-south-2"
  }

  assert {
    condition     = output.region == "eu-south-2"
    error_message = "eu-south-2 must be accepted."
  }
}

run "rejects_unknown_region" {
  command = plan

  variables {
    region = "us-east-1"
  }

  expect_failures = [var.region]
}

run "rejects_compliance_mode" {
  command = plan

  variables {
    evidence_object_lock = {
      mode = "COMPLIANCE"
      days = 1
    }
  }

  expect_failures = [var.evidence_object_lock]
}

run "api_service_is_composed" {
  command = plan

  assert {
    condition     = output.table_names.ephemeral == "lamassu-cla-staging-ephemeral" && output.github_client_secret_name == "lamassu-cla/staging/github-oauth-client-secret"
    error_message = "The root composes the F-07 API service for this environment."
  }
}

run "phase_a_without_artifact" {
  command = plan

  variables {
    api_package = null
  }

  assert {
    condition     = output.api_endpoint == null && output.github_client_secret_name == "lamassu-cla/staging/github-oauth-client-secret"
    error_message = "Without an artifact the root plans storage, keys, tables, and the secret only (A-32)."
  }
}
