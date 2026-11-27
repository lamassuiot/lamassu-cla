output "api_endpoint" {
  description = "Execute-api endpoint of the HTTP API; the CloudFront `/api/*` origin (D-02)."
  value       = module.http_api.api_endpoint
}

output "api_function_name" {
  description = "Name of the `api` function."
  value       = module.api.function_name
}

output "api_role_arn" {
  description = "Execution role of the `api` function."
  value       = module.api.role_arn
}

output "table_names" {
  description = "DynamoDB table names by data class."
  value = {
    core      = module.core_table.name
    audit     = module.audit_table.name
    ephemeral = module.ephemeral_table.name
  }
}

output "audit_stream_arn" {
  description = "Stream of the audit table, for the evidence export (ADR-0009)."
  value       = module.audit_table.stream_arn
}

output "github_client_secret_name" {
  description = "Name of the secret whose value the OAuth App owners set (D-30)."
  value       = module.github_client_secret.name
}

output "artifacts_bucket_name" {
  description = "Deployment-artifact bucket."
  value       = module.artifacts_bucket.bucket_name
}

output "api_environment" {
  description = "Environment variables of the `api` function, for tests. Configuration only; no secret values."
  value       = local.api_environment
}

output "api_data_policy" {
  description = "Data-access policy of the `api` function, for tests."
  value       = local.api_data_policy
}

output "operations_key_policy" {
  description = "Key policy of the operations key, for tests."
  value       = local.operations_key_policy
}
