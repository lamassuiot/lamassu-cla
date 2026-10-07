terraform {
  required_version = ">= 1.11"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.67.0, < 7.0.0"
    }
  }

  # Partial configuration: bucket, key, and region are supplied at `terraform init`.
  # The first apply uses a local backend override; see docs/operations/terraform-bootstrap.md.
  backend "s3" {
    use_lockfile = true
    encrypt      = true
  }
}
