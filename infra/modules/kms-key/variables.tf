variable "description" {
  description = "Purpose of the key, shown in the AWS console."
  type        = string
}

variable "alias_name" {
  description = "Key alias, including the `alias/` prefix."
  type        = string

  validation {
    condition     = startswith(var.alias_name, "alias/") && !startswith(var.alias_name, "alias/aws/")
    error_message = "alias_name must start with \"alias/\" and must not use the reserved \"alias/aws/\" prefix."
  }
}

variable "deletion_window_in_days" {
  description = "Waiting period before a scheduled key deletion completes."
  type        = number
  default     = 30

  validation {
    condition     = var.deletion_window_in_days >= 7 && var.deletion_window_in_days <= 30
    error_message = "deletion_window_in_days must be between 7 and 30."
  }
}

variable "policy" {
  description = "Key policy JSON. Null keeps the AWS default key policy, which delegates access to IAM in the account."
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
