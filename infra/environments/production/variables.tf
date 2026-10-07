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
