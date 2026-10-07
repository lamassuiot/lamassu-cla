provider "aws" {
  region = var.region

  default_tags {
    tags = {
      project      = "lamassu-cla"
      environment  = local.environment
      "managed-by" = "terraform"
    }
  }
}

locals {
  environment = "dev"
}
