output "role_arn" {
  description = "ARN of the IAM role."
  value       = aws_iam_role.this.arn
}

output "role_name" {
  description = "Name of the IAM role."
  value       = aws_iam_role.this.name
}

output "trust_policy" {
  description = "Trust policy JSON of the role."
  value       = local.trust_policy
}

output "inline_policies" {
  description = "Inline policy JSON documents attached to the role, keyed by name."
  value       = var.inline_policies
}
