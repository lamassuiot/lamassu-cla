mock_provider "aws" {}

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
