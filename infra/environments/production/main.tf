# Composition only (AC-03-2): environment roots call modules from infra/modules.

# F-07 sign-in API (unit 3, plan-only until F-03 step 5).
module "api_service" {
  source = "../../modules/api-service"

  environment           = local.environment
  portal_origin         = var.portal_origin
  github_client_id      = var.github_client_id
  sign_in_enabled       = var.sign_in_enabled
  artifacts_bucket_name = var.artifacts_bucket_name
  api_package           = var.api_package
  log_retention_in_days = var.log_retention_in_days
}
