locals {
  # Access-log fields: no query string, no client IP, no user agent, no cookies, no authorizer
  # claims (F-07 security considerations, security specification 7). The path carries no personal
  # data by contract (`return_to` and OAuth parameters travel in the query string).
  access_log_format = jsonencode({
    requestId               = "$context.requestId"
    requestTime             = "$context.requestTime"
    httpMethod              = "$context.httpMethod"
    path                    = "$context.path"
    routeKey                = "$context.routeKey"
    protocol                = "$context.protocol"
    status                  = "$context.status"
    responseLength          = "$context.responseLength"
    responseLatency         = "$context.responseLatency"
    integrationLatency      = "$context.integrationLatency"
    integrationStatus       = "$context.integrationStatus"
    integrationErrorMessage = "$context.integrationErrorMessage"
    errorMessage            = "$context.error.message"
    errorResponseType       = "$context.error.responseType"
  })
}

module "access_log_group" {
  source = "../log-group"

  name              = "/aws/apigateway/${var.name}"
  retention_in_days = var.log_retention_in_days
  kms_key_arn       = var.kms_key_arn
  tags              = var.tags
}

resource "aws_apigatewayv2_api" "this" {
  name          = var.name
  description   = var.description
  protocol_type = "HTTP"
  tags          = var.tags
}

resource "aws_apigatewayv2_integration" "lambda" {
  api_id                 = aws_apigatewayv2_api.this.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.lambda_invoke_arn
  integration_method     = "POST"
  payload_format_version = "2.0"
  timeout_milliseconds   = var.integration_timeout_milliseconds
}

resource "aws_apigatewayv2_route" "this" {
  # checkov:skip=CKV_AWS_309:The application authenticates every request itself with host-only session cookies, CSRF tokens, and the Origin check (F-07); the sign-in routes are public by design and no JWT or IAM authorizer applies to browser sessions.
  for_each = toset(var.route_keys)

  api_id             = aws_apigatewayv2_api.this.id
  route_key          = each.value
  target             = "integrations/${aws_apigatewayv2_integration.lambda.id}"
  authorization_type = "NONE"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = "$default"
  auto_deploy = true
  tags        = var.tags

  access_log_settings {
    destination_arn = module.access_log_group.arn
    format          = local.access_log_format
  }

  default_route_settings {
    throttling_burst_limit   = var.throttling_burst_limit
    throttling_rate_limit    = var.throttling_rate_limit
    detailed_metrics_enabled = true
  }
}

resource "aws_lambda_permission" "api" {
  statement_id  = "AllowInvokeFromHttpApi"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.this.execution_arn}/*/*"
}
