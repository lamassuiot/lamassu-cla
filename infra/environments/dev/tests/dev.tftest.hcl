mock_provider "aws" {}

run "plans_without_object_lock_values" {
  command = plan

  assert {
    condition     = output.environment == "dev" && output.region == "eu-west-1"
    error_message = "The dev root must default to eu-west-1."
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
