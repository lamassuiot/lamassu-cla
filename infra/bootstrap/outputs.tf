output "region" {
  description = "Region of the account resources. Set as the GitHub environment variable for deployments."
  value       = var.region
}

output "state_bucket_name" {
  description = "Terraform state bucket. Set as the GitHub environment variable for deployments."
  value       = module.state_bucket.bucket_name
}

output "state_kms_key_arn" {
  description = "KMS key that encrypts the state bucket."
  value       = module.state_key.key_arn
}

output "deployment_role_arn" {
  description = "Role assumed by deployment workflows through OIDC. Set as the GitHub environment variable."
  value       = module.deployment_role.role_arn
}

output "environment_state_key" {
  description = "Backend key for the environment root."
  value       = "${var.environment_state_prefix}terraform.tfstate"
}
