variable "function_name" {
  description = "Function name; also the IAM role name and the log group suffix."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9_-]{1,64}$", var.function_name))
    error_message = "function_name must be 1 to 64 letters, digits, hyphens, or underscores."
  }
}

variable "description" {
  description = "Function description."
  type        = string
}

variable "package" {
  description = "Deployment artifact in S3: a zip containing the `bootstrap` binary built by `make build-go`, uploaded by the deployment workflow. The object version pins the exact artifact."
  type = object({
    s3_bucket         = string
    s3_key            = string
    s3_object_version = string
  })

  validation {
    condition     = endswith(var.package.s3_key, ".zip") && length(var.package.s3_object_version) > 0
    error_message = "package.s3_key must point at a .zip object and package.s3_object_version must be set."
  }
}

variable "environment_variables" {
  description = "Environment variables. Configuration only: secrets travel as Secrets Manager identifiers, never as values."
  type        = map(string)
  default     = {}
}

variable "memory_size" {
  description = "Memory in MB (CPU scales with it)."
  type        = number
  default     = 512

  validation {
    condition     = var.memory_size >= 128 && var.memory_size <= 10240 && floor(var.memory_size) == var.memory_size
    error_message = "memory_size must be a whole number between 128 and 10240."
  }
}

variable "timeout" {
  description = "Timeout in seconds. API functions stay below the API Gateway integration limit of 30 seconds."
  type        = number
  default     = 10

  validation {
    condition     = var.timeout >= 1 && var.timeout <= 900 && floor(var.timeout) == var.timeout
    error_message = "timeout must be a whole number between 1 and 900."
  }
}

variable "reserved_concurrent_executions" {
  description = "Reserved concurrency; caps the function's share of the account (architecture 3.3). Never unlimited."
  type        = number

  validation {
    condition     = var.reserved_concurrent_executions >= 1 && floor(var.reserved_concurrent_executions) == var.reserved_concurrent_executions
    error_message = "reserved_concurrent_executions must be a whole number of at least 1."
  }
}

variable "application_log_level" {
  description = "Minimum application log level that reaches CloudWatch."
  type        = string
  default     = "INFO"

  validation {
    condition     = contains(["TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"], var.application_log_level)
    error_message = "application_log_level must be a Lambda application log level."
  }
}

variable "log_retention_in_days" {
  description = "Retention of the function's log group. No default (D-10)."
  type        = number
}

variable "kms_key_arn" {
  description = "ARN of the customer-managed KMS key that encrypts the environment variables and the log group. Its key policy must allow the CloudWatch Logs service."
  type        = string
}

variable "inline_policies" {
  description = "Additional inline policies for the execution role, by name, as policy document JSON. Only the actions and resources the function needs (security specification 9)."
  type        = map(string)
  default     = {}
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
