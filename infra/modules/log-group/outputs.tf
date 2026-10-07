output "name" {
  description = "Name of the log group."
  value       = aws_cloudwatch_log_group.this.name
}

output "arn" {
  description = "ARN of the log group."
  value       = aws_cloudwatch_log_group.this.arn
}
