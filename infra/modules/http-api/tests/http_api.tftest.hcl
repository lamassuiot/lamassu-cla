mock_provider "aws" {}

override_module {
  target = module.access_log_group
  outputs = {
    name = "/aws/apigateway/synthetic-api"
    arn  = "arn:aws:logs:eu-west-1:000000000000:log-group:/aws/apigateway/synthetic-api"
  }
}

variables {
  name                  = "synthetic-api"
  description           = "Synthetic test API"
  lambda_invoke_arn     = "arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:000000000000:function:synthetic-api/invocations"
  lambda_function_name  = "synthetic-api"
  route_keys            = ["ANY /api/{proxy+}", "GET /api"]
  log_retention_in_days = 30
  kms_key_arn           = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"
}

run "http_api_with_lambda_proxy" {
  command = plan

  assert {
    condition     = aws_apigatewayv2_api.this.protocol_type == "HTTP"
    error_message = "The API must be an HTTP API."
  }

  assert {
    condition     = aws_apigatewayv2_integration.lambda.integration_type == "AWS_PROXY" && aws_apigatewayv2_integration.lambda.payload_format_version == "2.0"
    error_message = "The function receives payload format 2.0 events (A-31)."
  }

  assert {
    condition     = aws_apigatewayv2_integration.lambda.timeout_milliseconds == 15000
    error_message = "The default integration timeout is 15 seconds."
  }

  assert {
    condition     = sort(keys(aws_apigatewayv2_route.this)) == sort(["ANY /api/{proxy+}", "GET /api"]) && alltrue([for r in aws_apigatewayv2_route.this : r.authorization_type == "NONE"])
    error_message = "Only the requested routes exist; the application authenticates sessions itself."
  }

  assert {
    condition     = aws_apigatewayv2_stage.default.name == "$default" && aws_apigatewayv2_stage.default.auto_deploy
    error_message = "The $default stage deploys automatically."
  }

  assert {
    condition     = one(aws_apigatewayv2_stage.default.default_route_settings).throttling_burst_limit == 100 && one(aws_apigatewayv2_stage.default.default_route_settings).throttling_rate_limit == 50
    error_message = "Stage throttling must be set."
  }

  assert {
    condition     = aws_lambda_permission.api.principal == "apigateway.amazonaws.com" && aws_lambda_permission.api.function_name == "synthetic-api"
    error_message = "Only API Gateway may invoke the function."
  }
}

run "access_logs_carry_no_personal_data" {
  command = plan

  assert {
    condition     = one(aws_apigatewayv2_stage.default.access_log_settings).destination_arn == "arn:aws:logs:eu-west-1:000000000000:log-group:/aws/apigateway/synthetic-api"
    error_message = "Access logs go to the encrypted, retention-bound log group."
  }

  assert {
    condition = length(setintersection(
      toset(values(jsondecode(output.access_log_format))),
      toset(["$context.identity.sourceIp", "$context.identity.userAgent", "$context.identity.user", "$context.identity.caller", "$context.authorizer.claims", "$context.requestOverride.querystring", "$context.querystring"])
    )) == 0
    error_message = "Access logs must not record the client IP, the user agent, identities, or query strings (F-07)."
  }

  assert {
    condition     = alltrue([for v in values(jsondecode(output.access_log_format)) : !strcontains(lower(v), "querystring") && !strcontains(v, "$context.identity") && !strcontains(v, "$input.")])
    error_message = "No access-log field may derive from the query string, the identity context, or the request body."
  }

  assert {
    condition     = contains(values(jsondecode(output.access_log_format)), "$context.requestId") && contains(values(jsondecode(output.access_log_format)), "$context.status")
    error_message = "Access logs keep the request ID and status for correlation."
  }
}

run "rejects_default_route" {
  command = plan

  variables {
    route_keys = ["$default"]
  }

  expect_failures = [var.route_keys]
}

run "rejects_timeout_above_limit" {
  command = plan

  variables {
    integration_timeout_milliseconds = 30001
  }

  expect_failures = [var.integration_timeout_milliseconds]
}
