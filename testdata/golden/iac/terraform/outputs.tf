# Root-module outputs. Surface anything a downstream consumer (another
# module, a CI pipeline, an Ansible inventory generator) might need.

output "environment" {
  description = "Environment short code, echoed back for sanity-checking."
  value       = var.environment
}
