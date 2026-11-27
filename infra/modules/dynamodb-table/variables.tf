variable "name" {
  description = "Table name."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9_.-]{3,255}$", var.name))
    error_message = "name must be 3 to 255 letters, digits, underscores, dots, or hyphens."
  }
}

variable "hash_key" {
  description = "Partition key attribute. Must appear in attributes."
  type        = string
}

variable "range_key" {
  description = "Sort key attribute. Must appear in attributes when set."
  type        = string
  default     = null
}

variable "attributes" {
  description = "Key attributes of the table and its indexes, by name, with their DynamoDB type (S, N, or B). Non-key attributes are not declared."
  type        = map(string)

  validation {
    condition     = alltrue([for t in values(var.attributes) : contains(["S", "N", "B"], t)])
    error_message = "Attribute types must be S, N, or B."
  }

  validation {
    condition     = contains(keys(var.attributes), var.hash_key) && (var.range_key == null || contains(keys(var.attributes), coalesce(var.range_key, "")))
    error_message = "hash_key and range_key must be declared in attributes."
  }
}

variable "global_secondary_indexes" {
  description = "Global secondary indexes by name. Every index projects all attributes; on-demand tables need no index throughput."
  type = map(object({
    hash_key  = string
    range_key = optional(string)
  }))
  default = {}

  validation {
    condition = alltrue([
      for i in values(var.global_secondary_indexes) :
      contains(keys(var.attributes), i.hash_key) && (i.range_key == null || contains(keys(var.attributes), coalesce(i.range_key, "")))
    ])
    error_message = "Index keys must be declared in attributes."
  }
}

variable "ttl_attribute" {
  description = "Numeric epoch-seconds attribute that expires items. Null disables TTL. Expiry is eventual; readers must still check the attribute (domain model 5.3)."
  type        = string
  default     = null
}

variable "stream_view_type" {
  description = "DynamoDB Streams view type (KEYS_ONLY, NEW_IMAGE, OLD_IMAGE, NEW_AND_OLD_IMAGES). Null disables the stream."
  type        = string
  default     = null

  validation {
    condition     = var.stream_view_type == null || contains(["KEYS_ONLY", "NEW_IMAGE", "OLD_IMAGE", "NEW_AND_OLD_IMAGES"], coalesce(var.stream_view_type, "KEYS_ONLY"))
    error_message = "stream_view_type must be a DynamoDB Streams view type or null."
  }
}

variable "kms_key_arn" {
  description = "ARN of the customer-managed KMS key that encrypts the table (domain model 5)."
  type        = string
}

variable "tags" {
  description = "Additional tags. Default tags come from the provider configuration."
  type        = map(string)
  default     = {}
}
