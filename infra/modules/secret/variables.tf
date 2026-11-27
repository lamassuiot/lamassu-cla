variable "name" {
  description = "Secret name."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9/_+=.@-]{1,512}$", var.name))
    error_message = "name must be 1 to 512 letters, digits, or /_+=.@- characters."
  }
}

variable "description" {
  description = "Secret description. Must not reveal the value or who holds it."
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of the customer-managed KMS key that encrypts the secret (security specification 4)."
  type        = string
}

variable "reader_role_arns" {
  description = "IAM role ARNs exempt from the deny on GetSecretValue; their identity policies grant the read. Every other principal, including administrators, is denied. The roles need not exist yet."
  type        = list(string)

  validation {
    condition     = length(var.reader_role_arns) >= 1 && alltrue([for a in var.reader_role_arns : can(regex("^arn:aws:iam::[0-9]{12}:role/", a))])
    error_message = "reader_role_arns must list at least one IAM role ARN."
  }
}

variable "recovery_window_in_days" {
  description = "Days before a deleted secret is removed for good. 0 is forbidden so a deletion can be reverted."
  type        = number
  default     = 30

  validation {
    condition     = var.recovery_window_in_days >= 7 && var.recovery_window_in_days <= 30 && floor(var.recovery_window_in_days) == var.recovery_window_in_days
    error_message = "recovery_window_in_days must be a whole number between 7 and 30."
  }
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
