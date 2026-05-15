// tflint baseline. Tighten per provider — `tflint --init` after editing.

config {
  call_module_type = "all"
  force            = false
  disabled_by_default = false
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

// Uncomment the providers you actually use; otherwise `tflint --init` will
// download every plugin every time.
//
// plugin "aws" {
//   enabled = true
//   version = "0.32.0"
//   source  = "github.com/terraform-linters/tflint-ruleset-aws"
// }
//
// plugin "google" {
//   enabled = true
//   version = "0.30.0"
//   source  = "github.com/terraform-linters/tflint-ruleset-google"
// }
//
// plugin "azurerm" {
//   enabled = true
//   version = "0.27.0"
//   source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
// }
