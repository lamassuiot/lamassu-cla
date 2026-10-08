variable "bucket_name" {
  description = "Globally unique bucket name. Supplied per environment; never committed."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", var.bucket_name))
    error_message = "bucket_name must be 3 to 63 lowercase letters, digits, dots, or hyphens."
  }
}

variable "kms_key_arn" {
  description = "ARN of the customer-managed KMS key used for default SSE-KMS encryption."
  type        = string
}

variable "noncurrent_version_expiration_days" {
  description = "Days after which non-current object versions expire. Null keeps every version."
  type        = number
  default     = null

  validation {
    condition     = var.noncurrent_version_expiration_days == null || coalesce(var.noncurrent_version_expiration_days, 0) >= 1
    error_message = "noncurrent_version_expiration_days must be null or at least 1."
  }
}

variable "abort_incomplete_multipart_upload_days" {
  description = "Days after which incomplete multipart uploads are aborted, in the lifecycle rule. Uploaded objects are not affected."
  type        = number
  default     = 7

  validation {
    condition     = var.abort_incomplete_multipart_upload_days >= 1 && floor(var.abort_incomplete_multipart_upload_days) == var.abort_incomplete_multipart_upload_days
    error_message = "abort_incomplete_multipart_upload_days must be a whole number of at least 1."
  }
}

variable "object_lock" {
  description = "Default Object Lock retention (ADR-0005). Null creates the bucket without Object Lock. Object Lock can only be enabled at creation."
  type = object({
    mode = string
    days = number
  })
  default = null

  validation {
    condition     = var.object_lock == null || contains(["GOVERNANCE", "COMPLIANCE"], try(var.object_lock.mode, ""))
    error_message = "object_lock.mode must be GOVERNANCE or COMPLIANCE."
  }

  validation {
    condition     = var.object_lock == null || try(var.object_lock.days >= 1 && floor(var.object_lock.days) == var.object_lock.days, false)
    error_message = "object_lock.days must be a whole number of at least 1."
  }
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
