# Deliberately insecure fixture for scripts/check-terraform-scan.test.sh. Never deploy.
terraform {
  required_version = ">= 1.11"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.67.0, < 7.0.0"
    }
  }
}

resource "aws_kms_key" "unrotated" {
  description         = "Synthetic insecure fixture"
  enable_key_rotation = false
}

resource "aws_cloudwatch_log_group" "unencrypted" {
  name = "synthetic-insecure-fixture"
}
