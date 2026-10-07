output "environment" {
  description = "Environment of this root module."
  value       = local.environment
}

output "region" {
  description = "Region of this environment."
  value       = var.region
}

output "evidence_object_lock" {
  description = "Object Lock settings for the evidence bucket, consumed when that bucket is added (ADR-0005)."
  value       = var.evidence_object_lock
}
