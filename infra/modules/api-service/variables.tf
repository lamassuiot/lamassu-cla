variable "environment" {
  description = "Environment name; prefixes every resource name."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "environment must be dev, staging, or production."
  }
}

variable "portal_origin" {
  description = "Portal origin, for example https://cla.example. The GitHub callback URL is derived from it (F-07)."
  type        = string

  validation {
    condition     = can(regex("^https://[a-z0-9.-]+(:[0-9]+)?$", var.portal_origin))
    error_message = "portal_origin must be an https origin without a path or trailing slash."
  }
}

variable "github_client_id" {
  description = "Client ID of the sign-in OAuth App (public identifier, D-30). The client secret is never a variable: its value is set in Secrets Manager by the OAuth App owners."
  type        = string

  validation {
    condition     = length(var.github_client_id) > 0
    error_message = "github_client_id must be set."
  }
}

variable "sign_in_enabled" {
  description = "Whether sign-in is enabled. Stays false until D-20 is decided; anything else is a reviewed change (F-07)."
  type        = bool
  default     = false
}

variable "sign_in_error_route" {
  description = "Portal route that shows sign-in errors."
  type        = string
  default     = "/sign-in/error"
}

variable "return_to_allowlist" {
  description = "Allowed `return_to` path prefixes (F-07 step 2)."
  type        = list(string)
  default     = ["/"]

  validation {
    condition     = length(var.return_to_allowlist) >= 1 && alltrue([for p in var.return_to_allowlist : startswith(p, "/") && !strcontains(p, ",")])
    error_message = "return_to_allowlist entries must be absolute paths without commas."
  }
}

variable "session_idle_seconds" {
  description = "Session idle timeout in seconds. Null keeps the application default (30 minutes, D-06)."
  type        = number
  default     = null
}

variable "session_absolute_seconds" {
  description = "Session absolute timeout in seconds. Null keeps the application default (8 hours, D-06)."
  type        = number
  default     = null
}

variable "artifacts_bucket_name" {
  description = "Globally unique name of the deployment-artifact bucket. Supplied per environment; never committed."
  type        = string
}

variable "artifact_noncurrent_version_expiration_days" {
  description = "Days after which superseded artifact versions expire."
  type        = number
  default     = 90
}

variable "api_package" {
  description = "Key and object version of the `api` zip in the artifact bucket, supplied by the deployment workflow."
  type = object({
    s3_key            = string
    s3_object_version = string
  })
}

variable "log_retention_in_days" {
  description = "Retention of the function and access-log groups. No default (D-10)."
  type        = number
}

variable "api_memory_size" {
  description = "Memory of the `api` function in MB."
  type        = number
  default     = 512
}

variable "api_timeout" {
  description = "Timeout of the `api` function in seconds."
  type        = number
  default     = 10
}

variable "api_reserved_concurrency" {
  description = "Reserved concurrency of the `api` function."
  type        = number
  default     = 20
}

variable "api_throttling_burst_limit" {
  description = "HTTP API burst limit (requests)."
  type        = number
  default     = 100
}

variable "api_throttling_rate_limit" {
  description = "HTTP API steady-state limit (requests per second)."
  type        = number
  default     = 50
}
