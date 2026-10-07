variable "environment" {
  description = "Environment that this AWS account hosts."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "environment must be dev, staging, or production."
  }
}

variable "region" {
  description = "AWS region. Only eu-west-1 and eu-south-2 are permitted (D-04)."
  type        = string
  default     = "eu-west-1"

  validation {
    condition     = contains(["eu-west-1", "eu-south-2"], var.region)
    error_message = "region must be eu-west-1 or eu-south-2 (D-04)."
  }
}

variable "github_repository" {
  description = "Repository whose deployment workflows may assume the role."
  type        = string
  default     = "lamassuiot/lamassu-cla"
}

variable "state_bucket_name" {
  description = "Name of the Terraform state bucket. Supplied by the platform owner; never committed."
  type        = string
}

variable "state_noncurrent_version_expiration_days" {
  description = "Days after which non-current state object versions expire. Chosen by the platform owner."
  type        = number
}

variable "environment_state_prefix" {
  description = "Key prefix under which environment roots store state. The deployment role is limited to it."
  type        = string
  default     = "environment/"

  validation {
    condition     = can(regex("^[a-z0-9-]+/$", var.environment_state_prefix)) && var.environment_state_prefix != "bootstrap/"
    error_message = "environment_state_prefix must be a single lowercase path segment ending in / and must not be bootstrap/."
  }
}

variable "create_github_oidc_provider" {
  description = "Create the GitHub Actions OIDC provider. Set to false when the account already has one."
  type        = bool
  default     = true
}
