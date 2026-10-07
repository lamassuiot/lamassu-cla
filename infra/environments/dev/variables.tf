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
  description = "Object Lock mode and default retention for the evidence bucket (ADR-0005). Outside production only governance mode is allowed; the period is supplied at deployment."
  type = object({
    mode = string
    days = number
  })
  default = null

  validation {
    condition     = var.evidence_object_lock == null || try(var.evidence_object_lock.mode, "") == "GOVERNANCE"
    error_message = "Outside production, evidence_object_lock.mode must be GOVERNANCE so test data can be cleaned up (ADR-0005)."
  }

  validation {
    condition     = var.evidence_object_lock == null || try(var.evidence_object_lock.days >= 1 && floor(var.evidence_object_lock.days) == var.evidence_object_lock.days, false)
    error_message = "evidence_object_lock.days must be a whole number of at least 1."
  }
}
