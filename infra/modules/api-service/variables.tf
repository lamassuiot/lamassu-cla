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
  description = "Key and object version of the `api` zip in the artifact bucket, supplied by the deployment workflow. Null plans only storage, keys, tables, and the secret (phase A of the first deployment, A-32)."
  type = object({
    s3_key            = string
    s3_object_version = string
  })
  default = null
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
  description = "Timeout of the `api` function in seconds. Must hold two upstream calls, the revocation budget, and a reserve for DynamoDB and the runtime (A-32)."
  type        = number
  default     = 15

  validation {
    condition     = var.api_timeout >= 2 * var.upstream_timeout_seconds + var.revoke_timeout_seconds + 3
    error_message = "api_timeout must be at least 2 x upstream_timeout_seconds + revoke_timeout_seconds + 3 seconds of reserve."
  }

  validation {
    condition     = var.api_timeout < var.api_integration_timeout_seconds
    error_message = "api_timeout must be below api_integration_timeout_seconds so the gateway reports the function's timeout, not its own."
  }
}

variable "api_integration_timeout_seconds" {
  description = "API Gateway integration timeout in seconds (at most 30)."
  type        = number
  default     = 20

  validation {
    condition     = var.api_integration_timeout_seconds >= 1 && var.api_integration_timeout_seconds <= 30
    error_message = "api_integration_timeout_seconds must be between 1 and 30."
  }
}

variable "upstream_timeout_seconds" {
  description = "Timeout of each GitHub call (token exchange, user lookup) in seconds; CLA_UPSTREAM_TIMEOUT."
  type        = number
  default     = 4

  validation {
    condition     = var.upstream_timeout_seconds >= 1
    error_message = "upstream_timeout_seconds must be at least 1."
  }
}

variable "revoke_timeout_seconds" {
  description = "Budget of the best-effort token revocation in seconds; CLA_REVOKE_TIMEOUT. It bounds how long revocation can delay a sign-in."
  type        = number
  default     = 2

  validation {
    condition     = var.revoke_timeout_seconds >= 1
    error_message = "revoke_timeout_seconds must be at least 1."
  }
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
