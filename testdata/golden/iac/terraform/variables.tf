# Root-module input variables. Each variable declared here is a knob that
# the caller (e.g. an environment-specific tfvars file or a CI pipeline)
# must or may set. Keep descriptions concrete — they show up in
# `terraform-docs` output.

variable "environment" {
  description = "Short environment code (dev | staging | prod). Drives resource naming."
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "tags" {
  description = "Tags applied to every taggable resource."
  type        = map(string)
  default     = {}
}
