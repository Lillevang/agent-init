# Root module for iac.
#
# Add resources here or, preferably, compose them out of modules under
# `terraform/modules/`. Wire inputs via `variables.tf` and surface outputs
# via `outputs.tf`. Provider/backend configuration lives in `versions.tf`.

# Example placeholder. Delete once you have real resources.
locals {
  project = "iac"
}

output "project_name" {
  value = local.project
}
