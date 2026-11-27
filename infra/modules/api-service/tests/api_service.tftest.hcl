mock_provider "aws" {
  mock_data "aws_caller_identity" {
    defaults = { account_id = "000000000000" }
  }

  mock_data "aws_region" {
    defaults = { region = "eu-west-1", name = "eu-west-1" }
  }
}

# Child modules are tested on their own; known outputs let plan-only runs inspect the policies
# composed from their ARNs.
override_module {
  target = module.core_table
  outputs = {
    name       = "lamassu-cla-dev-core"
    arn        = "arn:aws:dynamodb:eu-west-1:000000000000:table/lamassu-cla-dev-core"
    index_arns = {}
    stream_arn = null
  }
}

override_module {
  target = module.audit_table
  outputs = {
    name       = "lamassu-cla-dev-audit"
    arn        = "arn:aws:dynamodb:eu-west-1:000000000000:table/lamassu-cla-dev-audit"
    index_arns = {}
    stream_arn = "arn:aws:dynamodb:eu-west-1:000000000000:table/lamassu-cla-dev-audit/stream/2026-01-01T00:00:00.000"
  }
}

override_module {
  target = module.ephemeral_table
  outputs = {
    name       = "lamassu-cla-dev-ephemeral"
    arn        = "arn:aws:dynamodb:eu-west-1:000000000000:table/lamassu-cla-dev-ephemeral"
    index_arns = {}
    stream_arn = null
  }
}

override_module {
  target = module.secrets_key
  outputs = {
    key_arn    = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-00000000sec0"
    key_id     = "00000000-0000-0000-0000-00000000sec0"
    alias_name = "alias/lamassu-cla-dev-secrets"
  }
}

override_module {
  target = module.core_key
  outputs = {
    key_arn    = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-00000000c0re"
    key_id     = "00000000-0000-0000-0000-00000000c0re"
    alias_name = "alias/lamassu-cla-dev-core"
  }
}

override_module {
  target = module.audit_key
  outputs = {
    key_arn    = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-0000000aud1t"
    key_id     = "00000000-0000-0000-0000-0000000aud1t"
    alias_name = "alias/lamassu-cla-dev-audit"
  }
}

override_module {
  target = module.github_client_secret
  outputs = {
    arn             = "arn:aws:secretsmanager:eu-west-1:000000000000:secret:lamassu-cla/dev/github-oauth-client-secret-AbCdEf"
    name            = "lamassu-cla/dev/github-oauth-client-secret"
    resource_policy = "{}"
  }
}

override_module {
  target = module.api[0]
  outputs = {
    function_name  = "lamassu-cla-dev-api"
    function_arn   = "arn:aws:lambda:eu-west-1:000000000000:function:lamassu-cla-dev-api"
    invoke_arn     = "arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:000000000000:function:lamassu-cla-dev-api/invocations"
    role_arn       = "arn:aws:iam::000000000000:role/lamassu-cla-dev-api"
    role_name      = "lamassu-cla-dev-api"
    log_group_name = "/aws/lambda/lamassu-cla-dev-api"
    logging_policy = "{}"
  }
}

variables {
  environment           = "dev"
  portal_origin         = "https://cla.example"
  github_client_id      = "synthetic-client-id"
  artifacts_bucket_name = "synthetic-artifacts"
  log_retention_in_days = 30

  api_package = {
    s3_key            = "api/0123456789abcdef.zip"
    s3_object_version = "synthetic-version"
  }
}

run "composition" {
  command = plan

  assert {
    condition     = output.table_names == { core = "lamassu-cla-dev-core", audit = "lamassu-cla-dev-audit", ephemeral = "lamassu-cla-dev-ephemeral" }
    error_message = "Three tables per environment (ADR-0009)."
  }

  assert {
    condition     = output.api_function_name == "lamassu-cla-dev-api" && output.github_client_secret_name == "lamassu-cla/dev/github-oauth-client-secret"
    error_message = "Names are prefixed with the environment."
  }

  assert {
    condition     = output.audit_stream_arn != null && module.api[0].function_name == "lamassu-cla-dev-api"
    error_message = "The audit table streams for export."
  }

  assert {
    condition     = module.http_api[0].access_log_group_name == "/aws/apigateway/lamassu-cla-dev-api"
    error_message = "The HTTP API carries the function name."
  }

  assert {
    condition     = output.api_role_arn == "arn:aws:iam::000000000000:role/lamassu-cla-dev-api"
    error_message = "The role ARN is derived from the function name, so the secret can name it before it exists."
  }
}

run "phase_a_without_artifact" {
  command = plan

  variables {
    api_package = null
  }

  assert {
    condition     = length(module.api) == 0 && length(module.http_api) == 0 && output.api_endpoint == null && output.api_function_name == null
    error_message = "Without an artifact no function or HTTP API is planned (first-deployment sequence, A-32)."
  }

  assert {
    condition     = output.github_client_secret_name == "lamassu-cla/dev/github-oauth-client-secret" && output.table_names.core == "lamassu-cla-dev-core"
    error_message = "Storage, keys, tables, and the secret are planned in phase A."
  }
}

run "api_environment_is_configuration_only" {
  command = plan

  assert {
    condition = output.api_environment == {
      CLA_PORTAL_ORIGIN           = "https://cla.example"
      CLA_GITHUB_CALLBACK_URL     = "https://cla.example/api/auth/github/callback"
      CLA_SIGN_IN_ERROR_ROUTE     = "/sign-in/error"
      CLA_SIGN_IN_ENABLED         = "false"
      CLA_RETURN_TO_ALLOWLIST     = "/"
      CLA_TABLE_EPHEMERAL         = "lamassu-cla-dev-ephemeral"
      CLA_TABLE_CORE              = "lamassu-cla-dev-core"
      CLA_TABLE_AUDIT             = "lamassu-cla-dev-audit"
      CLA_GITHUB_CLIENT_ID        = "synthetic-client-id"
      CLA_GITHUB_CLIENT_SECRET_ID = "lamassu-cla/dev/github-oauth-client-secret"
      CLA_UPSTREAM_TIMEOUT        = "4"
      CLA_REVOKE_TIMEOUT          = "2"
    }
    error_message = "The function gets exactly the cmd/api variables; sign-in is off; the secret travels as an identifier; the budgets are explicit."
  }

  assert {
    condition     = !contains(keys(output.api_environment), "CLA_GITHUB_CLIENT_SECRET") && !contains(keys(output.api_environment), "CLA_DYNAMODB_ENDPOINT")
    error_message = "No secret value and no local endpoint in the deployed environment."
  }
}

run "session_overrides" {
  command = plan

  variables {
    sign_in_enabled          = true
    session_idle_seconds     = 600
    session_absolute_seconds = 3600
    return_to_allowlist      = ["/", "/agreements/"]
  }

  assert {
    condition     = output.api_environment["CLA_SIGN_IN_ENABLED"] == "true" && output.api_environment["CLA_SESSION_IDLE"] == "600" && output.api_environment["CLA_SESSION_ABSOLUTE"] == "3600" && output.api_environment["CLA_RETURN_TO_ALLOWLIST"] == "/,/agreements/"
    error_message = "Session policy and allowlist overrides reach the function."
  }
}

run "least_privilege_data_access" {
  command = plan

  assert {
    condition     = [for s in jsondecode(output.api_data_policy).Statement : s.Sid] == ["EphemeralAttemptsAndSessions", "CoreContributorProfile", "AuditAppend", "AuditAppendOnly", "DecryptTableKeys", "ReadGitHubClientSecret", "DecryptSecrets"]
    error_message = "The data policy has exactly the F-07 statements."
  }

  assert {
    condition     = alltrue([for s in jsondecode(output.api_data_policy).Statement : s.Resource != "*" && !strcontains(jsonencode(s.Action), "*")])
    error_message = "No wildcard actions or resources (security specification 9)."
  }

  assert {
    condition     = one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "AuditAppend"]).Action == "dynamodb:PutItem"
    error_message = "The audit table only receives PutItem."
  }

  assert {
    condition     = one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "AuditAppendOnly"]).Effect == "Deny" && contains(one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "AuditAppendOnly"]).Action, "dynamodb:DeleteItem") && contains(one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "AuditAppendOnly"]).Action, "dynamodb:UpdateItem")
    error_message = "Audit updates and deletes are explicitly denied (ADR-0009)."
  }

  assert {
    condition     = !contains(one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "CoreContributorProfile"]).Action, "dynamodb:DeleteItem") && !contains(one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "CoreContributorProfile"]).Action, "dynamodb:Scan")
    error_message = "The sign-in function cannot delete or scan core data."
  }

  assert {
    condition     = one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "DecryptSecrets"]).Condition.StringEquals["kms:ViaService"] == "secretsmanager.eu-west-1.amazonaws.com"
    error_message = "The secrets key is usable only through Secrets Manager."
  }

  assert {
    condition     = one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "DecryptTableKeys"]).Action == ["kms:Decrypt", "kms:DescribeKey"] && length(one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "DecryptTableKeys"]).Resource) == 2
    error_message = "The function may only decrypt the core and audit table keys; grants and data-key generation stay with the table creator."
  }

  assert {
    condition     = one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "DecryptTableKeys"]).Condition.StringEquals["kms:ViaService"] == "dynamodb.eu-west-1.amazonaws.com" && one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "DecryptTableKeys"]).Condition.StringEquals["kms:EncryptionContext:aws:dynamodb:subscriberId"] == "000000000000" && sort(one([for s in jsondecode(output.api_data_policy).Statement : s if s.Sid == "DecryptTableKeys"]).Condition.StringEquals["kms:EncryptionContext:aws:dynamodb:tableName"]) == tolist(["lamassu-cla-dev-audit", "lamassu-cla-dev-core", "lamassu-cla-dev-ephemeral"])
    error_message = "Table-key decryption is bound to DynamoDB, this account, and these three tables."
  }
}

run "timeout_budget" {
  command = plan

  assert {
    condition     = var.api_timeout >= 2 * var.upstream_timeout_seconds + var.revoke_timeout_seconds + 3 && var.api_timeout < var.api_integration_timeout_seconds
    error_message = "Defaults: 2 x 4 s upstream + 2 s revocation + 3 s reserve <= 15 s function < 20 s integration."
  }
}

run "rejects_budget_overrun" {
  command = plan

  variables {
    upstream_timeout_seconds = 6
  }

  expect_failures = [var.api_timeout]
}

run "rejects_function_timeout_at_integration_timeout" {
  command = plan

  variables {
    api_timeout                     = 20
    api_integration_timeout_seconds = 20
  }

  expect_failures = [var.api_timeout]
}

run "operations_key_serves_only_this_environment" {
  command = plan

  assert {
    condition     = one([for s in jsondecode(output.operations_key_policy).Statement : s if s.Sid == "CloudWatchLogs"]).Condition.ArnLike["kms:EncryptionContext:aws:logs:arn"] == "arn:aws:logs:eu-west-1:000000000000:log-group:/aws/*/lamassu-cla-dev-*"
    error_message = "CloudWatch Logs may use the key only for this environment's log groups."
  }
}

run "rejects_http_portal_origin" {
  command = plan

  variables {
    portal_origin = "http://cla.example"
  }

  expect_failures = [var.portal_origin]
}

run "rejects_relative_return_to" {
  command = plan

  variables {
    return_to_allowlist = ["agreements"]
  }

  expect_failures = [var.return_to_allowlist]
}
