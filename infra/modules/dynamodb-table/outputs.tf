output "name" {
  description = "Table name."
  value       = aws_dynamodb_table.this.name
}

output "arn" {
  description = "Table ARN."
  value       = aws_dynamodb_table.this.arn
}

output "index_arns" {
  description = "ARNs of the global secondary indexes, by index name."
  value       = { for name in keys(var.global_secondary_indexes) : name => "${aws_dynamodb_table.this.arn}/index/${name}" }
}

output "stream_arn" {
  description = "Stream ARN, or null when the stream is disabled."
  value       = var.stream_view_type == null ? null : aws_dynamodb_table.this.stream_arn
}
