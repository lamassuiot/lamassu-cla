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

output "api_endpoint" {
  description = "Execute-api endpoint of the HTTP API; the CloudFront `/api/*` origin (D-02)."
  value       = module.api_service.api_endpoint
}

output "github_client_secret_name" {
  description = "Secret whose value the OAuth App owners set (D-30)."
  value       = module.api_service.github_client_secret_name
}

output "table_names" {
  description = "DynamoDB table names by data class."
  value       = module.api_service.table_names
}
