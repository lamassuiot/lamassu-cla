terraform {
  required_version = ">= 1.11"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.67.0, < 7.0.0"
    }
  }

  # Partial configuration: bucket, key, and region come from GitHub environment variables at
  # `terraform init` (D-05). Nothing environment-specific is committed.
  backend "s3" {
    use_lockfile = true
    encrypt      = true
  }
}
