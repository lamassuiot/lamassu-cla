locals {
  log_group_name = "/aws/lambda/${var.function_name}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "LambdaService"
        Effect    = "Allow"
        Principal = { Service = "lambda.amazonaws.com" }
        Action    = "sts:AssumeRole"
      },
    ]
  })

  # Writing to its own log group only, instead of AWSLambdaBasicExecutionRole's `*` resource.
  logging_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "WriteOwnLogs"
        Effect   = "Allow"
        Action   = ["logs:CreateLogStream", "logs:PutLogEvents"]
        Resource = "${module.log_group.arn}:*"
      },
      {
        Sid      = "EncryptEnvironment"
        Effect   = "Allow"
        Action   = ["kms:Decrypt"]
        Resource = var.kms_key_arn
      },
    ]
  })
}

module "log_group" {
  source = "../log-group"

  name              = local.log_group_name
  retention_in_days = var.log_retention_in_days
  kms_key_arn       = var.kms_key_arn
  tags              = var.tags
}

resource "aws_iam_role" "this" {
  name               = var.function_name
  assume_role_policy = local.assume_role_policy
  tags               = var.tags
}

resource "aws_iam_role_policy" "logging" {
  name   = "logging"
  role   = aws_iam_role.this.id
  policy = local.logging_policy
}

resource "aws_iam_role_policy" "this" {
  for_each = var.inline_policies

  name   = each.key
  role   = aws_iam_role.this.id
  policy = each.value
}

resource "aws_lambda_function" "this" {
  # checkov:skip=CKV_AWS_116:Synchronous API functions have no asynchronous invocations to dead-letter; queue-driven workers get queues and DLQs with D-17.
  # checkov:skip=CKV_AWS_117:The function reaches only AWS APIs and GitHub over the public internet; no VPC resources exist (architecture 3.3).
  # checkov:skip=CKV_AWS_272:Code signing is not part of the approved deployment design; artifacts are built and uploaded by the GitHub OIDC deployment role (F-03).
  # checkov:skip=CKV_AWS_50:X-Ray tracing is decided with the observability design (alarms, dashboards; A-32); it would need xray:* permissions on every resource.
  function_name = var.function_name
  description   = var.description
  role          = aws_iam_role.this.arn

  # Go custom runtime (architecture 2, ADR-0008): the artifact is a zip with a `bootstrap` binary.
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  handler       = "bootstrap"
  package_type  = "Zip"

  s3_bucket         = var.package.s3_bucket
  s3_key            = var.package.s3_key
  s3_object_version = var.package.s3_object_version

  memory_size                    = var.memory_size
  timeout                        = var.timeout
  reserved_concurrent_executions = var.reserved_concurrent_executions
  kms_key_arn                    = var.kms_key_arn
  tags                           = var.tags

  environment {
    variables = var.environment_variables
  }

  logging_config {
    log_format            = "JSON"
    log_group             = module.log_group.name
    application_log_level = var.application_log_level
    system_log_level      = "WARN"
  }

  depends_on = [aws_iam_role_policy.logging]
}
