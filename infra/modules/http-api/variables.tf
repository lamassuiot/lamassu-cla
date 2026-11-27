variable "name" {
  description = "API name; also the access-log group suffix."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9_-]{1,128}$", var.name))
    error_message = "name must be 1 to 128 letters, digits, hyphens, or underscores."
  }
}

variable "description" {
  description = "API description."
  type        = string
}

variable "lambda_invoke_arn" {
  description = "Invoke ARN of the function that serves every route."
  type        = string
}

variable "lambda_function_name" {
  description = "Name of that function, for the invoke permission."
  type        = string
}

variable "route_keys" {
  description = "HTTP API route keys served by the function, for example `ANY /api/{proxy+}`. No `$default` route: unknown paths get 404 from API Gateway without invoking the function."
  type        = list(string)

  validation {
    condition     = length(var.route_keys) >= 1 && alltrue([for k in var.route_keys : can(regex("^(ANY|GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) /", k))])
    error_message = "route_keys must be `METHOD /path` entries; the $default route is not allowed."
  }
}

variable "integration_timeout_milliseconds" {
  description = "Integration timeout. HTTP APIs allow at most 30000."
  type        = number
  default     = 15000

  validation {
    condition     = var.integration_timeout_milliseconds >= 50 && var.integration_timeout_milliseconds <= 30000
    error_message = "integration_timeout_milliseconds must be between 50 and 30000."
  }
}

variable "throttling_burst_limit" {
  description = "Stage-level burst limit (requests)."
  type        = number
  default     = 100

  validation {
    condition     = var.throttling_burst_limit >= 1
    error_message = "throttling_burst_limit must be at least 1."
  }
}

variable "throttling_rate_limit" {
  description = "Stage-level steady-state limit (requests per second)."
  type        = number
  default     = 50

  validation {
    condition     = var.throttling_rate_limit >= 1
    error_message = "throttling_rate_limit must be at least 1."
  }
}

variable "log_retention_in_days" {
  description = "Retention of the access-log group. No default (D-10)."
  type        = number
}

variable "kms_key_arn" {
  description = "ARN of the customer-managed KMS key that encrypts the access-log group. Its key policy must allow the CloudWatch Logs service."
  type        = string
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
