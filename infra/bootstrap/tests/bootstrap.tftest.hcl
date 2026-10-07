mock_provider "aws" {
  override_during = plan

  mock_resource "aws_iam_openid_connect_provider" {
    defaults = {
      arn = "arn:aws:iam::000000000000:oidc-provider/token.actions.githubusercontent.com"
    }
  }

  mock_resource "aws_kms_key" {
    defaults = {
      arn = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"
    }
  }

  mock_resource "aws_s3_bucket" {
    defaults = {
      arn = "arn:aws:s3:::synthetic-state-bucket"
      id  = "synthetic-state-bucket"
    }
  }
}

variables {
  environment                              = "staging"
  state_bucket_name                        = "synthetic-state-bucket"
  state_noncurrent_version_expiration_days = 1
}

run "oidc_role_is_restricted_to_repository_and_environment" {
  command = plan

  assert {
    condition     = jsondecode(module.deployment_role.trust_policy).Statement[0].Condition.StringEquals["token.actions.githubusercontent.com:sub"] == "repo:lamassuiot/lamassu-cla:environment:staging"
    error_message = "The role must trust only this repository and the named GitHub environment."
  }

  assert {
    condition     = jsondecode(module.deployment_role.trust_policy).Statement[0].Condition.StringEquals["token.actions.githubusercontent.com:aud"] == "sts.amazonaws.com"
    error_message = "The audience must be sts.amazonaws.com."
  }

  assert {
    condition     = aws_iam_openid_connect_provider.github[0].client_id_list == toset(["sts.amazonaws.com"])
    error_message = "The OIDC provider must accept only the sts.amazonaws.com audience."
  }
}

run "role_is_limited_to_environment_state" {
  command = plan

  assert {
    condition = toset(flatten([
      for s in jsondecode(module.deployment_role.inline_policies["terraform-state"]).Statement : s.Action
    ])) == toset(["s3:ListBucket", "s3:GetObject", "s3:PutObject", "s3:DeleteObject", "kms:Decrypt", "kms:Encrypt", "kms:GenerateDataKey"])
    error_message = "The deployment role may only read and write state objects and use the state key."
  }

  assert {
    condition = [
      for s in jsondecode(module.deployment_role.inline_policies["terraform-state"]).Statement : s.Resource
      if s.Sid == "ReadWriteEnvironmentState"
    ] == ["arn:aws:s3:::synthetic-state-bucket/environment/*"]
    error_message = "Object access must be limited to the environment state prefix."
  }

  assert {
    condition = [
      for s in jsondecode(module.deployment_role.inline_policies["terraform-state"]).Statement : s.Condition.StringLike["s3:prefix"]
      if s.Sid == "ListEnvironmentState"
    ] == [["environment/*"]]
    error_message = "Listing must be limited to the environment state prefix."
  }

  assert {
    condition     = output.environment_state_key == "environment/terraform.tfstate"
    error_message = "The environment state key must be under the environment prefix."
  }
}

run "defaults_to_eu_west_1" {
  command = plan

  assert {
    condition     = output.region == "eu-west-1"
    error_message = "The default region must be eu-west-1."
  }
}

run "rejects_unknown_region" {
  command = plan

  variables {
    region = "us-east-1"
  }

  expect_failures = [var.region]
}

run "rejects_bootstrap_prefix_for_environments" {
  command = plan

  variables {
    environment_state_prefix = "bootstrap/"
  }

  expect_failures = [var.environment_state_prefix]
}
