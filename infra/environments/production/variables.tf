variable "region" {
  description = "AWS region. Only eu-west-1 and eu-south-2 are permitted (D-04)."
  type        = string
  default     = "eu-west-1"

  validation {
    condition     = contains(["eu-west-1", "eu-south-2"], var.region)
    error_message = "region must be eu-west-1 or eu-south-2 (D-04)."
  }
}

variable "evidence_object_lock" {
  description = "Object Lock mode and default retention for the evidence bucket (ADR-0005). Production has no usable default: Legal must confirm the values (D-10)."
  type = object({
    mode = string
    days = number
  })
  default = null

  validation {
    condition     = var.evidence_object_lock != null
    error_message = "Production Object Lock mode and retention are not set. They require Legal confirmation (D-10) and are supplied at deployment; no default exists."
  }

  validation {
    condition     = var.evidence_object_lock == null || contains(["GOVERNANCE", "COMPLIANCE"], try(var.evidence_object_lock.mode, ""))
    error_message = "evidence_object_lock.mode must be GOVERNANCE or COMPLIANCE."
  }

  validation {
    condition     = var.evidence_object_lock == null || try(var.evidence_object_lock.days >= 1 && floor(var.evidence_object_lock.days) == var.evidence_object_lock.days, false)
    error_message = "evidence_object_lock.days must be a whole number of at least 1."
  }
}

# F-07 sign-in API (unit 3). Environment-specific values come from GitHub environment variables
# at plan time (D-05); nothing environment-specific is committed.

variable "portal_origin" {
  description = "Portal origin of this environment (https, no path). Custom domains per environment are pending D-16."
  type        = string
}

variable "github_client_id" {
  description = "Client ID of the sign-in OAuth App for this environment (D-30). Its client secret is set in Secrets Manager by the OAuth App owners, never here."
  type        = string
}

variable "sign_in_enabled" {
  description = "Whether sign-in is enabled. Production sign-in stays disabled until D-20 is decided; the pull request that records D-20 removes this guard."
  type        = bool
  default     = false

  validation {
    condition     = !var.sign_in_enabled
    error_message = "Production sign-in stays disabled until D-20 is decided (F-07)."
  }
}

variable "artifacts_bucket_name" {
  description = "Globally unique name of the deployment-artifact bucket. Supplied per environment; never committed."
  type        = string
}

variable "api_package" {
  description = "Key and object version of the `api` zip uploaded by the deployment workflow."
  type = object({
    s3_key            = string
    s3_object_version = string
  })
}

variable "log_retention_in_days" {
  description = "Retention of the API log groups. No default: log retention is a Legal and DPO decision (D-10)."
  type        = number
}

