variable "billing_account" {
  description = "The ID of the billing account to associate projects with"
  type        = string
  default     = "AAAAAA-AAAAAA-AAAAAA"
}

variable "org_id" {
  description = "The organization id for the associated resources"
  type        = string
  default     = "123456789012"
}
