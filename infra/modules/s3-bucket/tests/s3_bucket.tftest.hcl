mock_provider "aws" {
  override_during = plan

  mock_resource "aws_s3_bucket" {
    defaults = {
      arn = "arn:aws:s3:::synthetic-test-bucket"
      id  = "synthetic-test-bucket"
    }
  }
}

variables {
  bucket_name = "synthetic-test-bucket"
  kms_key_arn = "arn:aws:kms:eu-west-1:000000000000:key/00000000-0000-0000-0000-000000000000"
}

run "secure_defaults" {
  command = plan

  assert {
    condition = alltrue([
      aws_s3_bucket_public_access_block.this.block_public_acls,
      aws_s3_bucket_public_access_block.this.block_public_policy,
      aws_s3_bucket_public_access_block.this.ignore_public_acls,
      aws_s3_bucket_public_access_block.this.restrict_public_buckets,
    ])
    error_message = "Every bucket-level Block Public Access setting must be enabled."
  }

  assert {
    condition     = aws_s3_bucket_versioning.this.versioning_configuration[0].status == "Enabled"
    error_message = "Versioning must be enabled."
  }

  assert {
    condition = one([
      for r in aws_s3_bucket_server_side_encryption_configuration.this.rule :
      one(r.apply_server_side_encryption_by_default).sse_algorithm
    ]) == "aws:kms"
    error_message = "Default encryption must be SSE-KMS."
  }

  assert {
    condition     = aws_s3_bucket_ownership_controls.this.rule[0].object_ownership == "BucketOwnerEnforced"
    error_message = "ACLs must be disabled with BucketOwnerEnforced."
  }

  assert {
    condition     = jsondecode(aws_s3_bucket_policy.this.policy).Statement[0].Condition.Bool["aws:SecureTransport"] == "false"
    error_message = "The bucket policy must deny requests without TLS."
  }

  assert {
    condition     = jsondecode(aws_s3_bucket_policy.this.policy).Statement[0].Effect == "Deny"
    error_message = "The TLS statement must deny."
  }

  assert {
    condition     = !aws_s3_bucket.this.force_destroy
    error_message = "force_destroy must be disabled."
  }

  assert {
    condition     = length(aws_s3_bucket_object_lock_configuration.this) == 0 && length(aws_s3_bucket_lifecycle_configuration.this) == 0
    error_message = "Object Lock and lifecycle rules must be opt-in."
  }
}

run "object_lock_and_lifecycle" {
  command = plan

  variables {
    object_lock = {
      mode = "GOVERNANCE"
      days = 1
    }
    noncurrent_version_expiration_days = 1
  }

  assert {
    condition     = aws_s3_bucket.this.object_lock_enabled
    error_message = "Object Lock must be enabled when configured."
  }

  assert {
    condition     = aws_s3_bucket_object_lock_configuration.this[0].rule[0].default_retention[0].mode == "GOVERNANCE"
    error_message = "The configured Object Lock mode must be applied."
  }

  assert {
    condition     = length(aws_s3_bucket_lifecycle_configuration.this) == 1
    error_message = "The non-current version expiry rule must be created when configured."
  }
}

run "rejects_invalid_object_lock_mode" {
  command = plan

  variables {
    object_lock = {
      mode = "NONE"
      days = 1
    }
  }

  expect_failures = [var.object_lock]
}

run "rejects_uppercase_bucket_name" {
  command = plan

  variables {
    bucket_name = "Synthetic-Test-Bucket"
  }

  expect_failures = [var.bucket_name]
}
