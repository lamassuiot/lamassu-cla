variable "role_name" {
  description = "Name of the IAM role."
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the GitHub Actions OIDC identity provider in the account."
  type        = string
}

variable "github_repository" {
  description = "Repository allowed to assume the role, as `owner/name`."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$", var.github_repository))
    error_message = "github_repository must have the form owner/name, without wildcards."
  }
}

variable "github_environment" {
  description = "GitHub environment allowed to assume the role."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "production"], var.github_environment)
    error_message = "github_environment must be dev, staging, or production."
  }
}

variable "inline_policies" {
  description = "Inline policy documents (JSON) keyed by policy name. Grant least privilege."
  type        = map(string)
  default     = {}
}

variable "max_session_duration" {
  description = "Maximum session duration in seconds."
  type        = number
  default     = 3600

  validation {
    condition     = var.max_session_duration >= 900 && var.max_session_duration <= 3600
    error_message = "max_session_duration must be between 900 and 3600 seconds."
  }
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
