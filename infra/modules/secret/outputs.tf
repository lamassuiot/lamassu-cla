output "arn" {
  description = "Secret ARN."
  value       = aws_secretsmanager_secret.this.arn
}

output "name" {
  description = "Secret name, usable as the identifier the application passes to Secrets Manager."
  value       = aws_secretsmanager_secret.this.name
}

output "resource_policy" {
  description = "Resource policy JSON, for tests."
  value       = local.resource_policy
}
