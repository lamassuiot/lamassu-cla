output "api_id" {
  description = "API ID."
  value       = aws_apigatewayv2_api.this.id
}

output "api_endpoint" {
  description = "Default execute-api endpoint. CloudFront uses it as the `/api/*` origin (D-02)."
  value       = aws_apigatewayv2_api.this.api_endpoint
}

output "execution_arn" {
  description = "Execution ARN of the API."
  value       = aws_apigatewayv2_api.this.execution_arn
}

output "access_log_group_name" {
  description = "Access-log group name."
  value       = module.access_log_group.name
}

output "access_log_format" {
  description = "Access-log format JSON, for tests."
  value       = local.access_log_format
}
