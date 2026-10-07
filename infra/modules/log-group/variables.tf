variable "name" {
  description = "Name of the CloudWatch log group."
  type        = string
}

variable "retention_in_days" {
  description = "Log retention in days. No default: log retention is a Legal and DPO decision (D-10)."
  type        = number

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653], var.retention_in_days)
    error_message = "retention_in_days must be a CloudWatch Logs retention value; never-expire (0) is not allowed."
  }
}

variable "kms_key_arn" {
  description = "ARN of the customer-managed KMS key that encrypts the log group. Its key policy must allow the CloudWatch Logs service."
  type        = string
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
