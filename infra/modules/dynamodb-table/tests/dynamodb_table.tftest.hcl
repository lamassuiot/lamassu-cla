mock_provider "aws" {
  override_during = plan
}

variables {
  name        = "synthetic-core"
  hash_key    = "PK"
  range_key   = "SK"
  attributes  = { PK = "S", SK = "S", GSI1PK = "S" }
  kms_key_arn = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"

  global_secondary_indexes = {
    GSI1 = { hash_key = "GSI1PK" }
  }
}

run "secure_defaults" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.this.billing_mode == "PAY_PER_REQUEST"
    error_message = "Tables must use on-demand capacity (domain model 5)."
  }

  assert {
    condition     = aws_dynamodb_table.this.deletion_protection_enabled
    error_message = "Deletion protection must be enabled."
  }

  assert {
    condition     = one(aws_dynamodb_table.this.point_in_time_recovery).enabled
    error_message = "Point-in-time recovery must be enabled."
  }

  assert {
    condition     = one(aws_dynamodb_table.this.server_side_encryption).enabled && one(aws_dynamodb_table.this.server_side_encryption).kms_key_arn == var.kms_key_arn
    error_message = "The table must be encrypted with the customer-managed key."
  }

  assert {
    condition     = length(aws_dynamodb_table.this.ttl) == 0 && !aws_dynamodb_table.this.stream_enabled
    error_message = "TTL and streams are off unless requested."
  }

  assert {
    condition     = one(aws_dynamodb_table.this.global_secondary_index).projection_type == "ALL" && one(aws_dynamodb_table.this.global_secondary_index).name == "GSI1"
    error_message = "Indexes project all attributes."
  }

  assert {
    condition     = [for k in one(aws_dynamodb_table.this.global_secondary_index).key_schema : "${k.key_type}=${k.attribute_name}"] == ["HASH=GSI1PK"]
    error_message = "Index keys follow the requested schema."
  }

  assert {
    condition     = output.stream_arn == null
    error_message = "No stream ARN without a stream."
  }
}

run "ttl_and_stream" {
  command = plan

  variables {
    name          = "synthetic-ephemeral"
    range_key     = null
    attributes    = { PK = "S", GSI1PK = "S", GSI1SK = "S" }
    ttl_attribute = "expires_at"
    global_secondary_indexes = {
      GSI1 = { hash_key = "GSI1PK", range_key = "GSI1SK" }
    }
    stream_view_type = "NEW_AND_OLD_IMAGES"
  }

  assert {
    condition     = [for k in one(aws_dynamodb_table.this.global_secondary_index).key_schema : "${k.key_type}=${k.attribute_name}"] == ["HASH=GSI1PK", "RANGE=GSI1SK"]
    error_message = "A sorted index declares both keys."
  }

  assert {
    condition     = one(aws_dynamodb_table.this.ttl).attribute_name == "expires_at" && one(aws_dynamodb_table.this.ttl).enabled
    error_message = "TTL must use the requested attribute."
  }

  assert {
    condition     = aws_dynamodb_table.this.stream_enabled && aws_dynamodb_table.this.stream_view_type == "NEW_AND_OLD_IMAGES"
    error_message = "The stream must use the requested view type."
  }
}

run "rejects_undeclared_key" {
  command = plan

  variables {
    attributes = { PK = "S", SK = "S" }
  }

  expect_failures = [var.global_secondary_indexes]
}

run "rejects_unknown_attribute_type" {
  command = plan

  variables {
    attributes = { PK = "S", SK = "L", GSI1PK = "S" }
  }

  expect_failures = [var.attributes]
}
