terraform {
  required_version = ">= 1.6.0, < 2.0.0"

  # Uncomment the providers you actually use and pin them. Wide constraints
  // ("anything >= X") are a rollback hazard once a new major drops.
  required_providers {
    # aws = {
    #   source  = "hashicorp/aws"
    #   version = ">= 5.50.0, < 6.0.0"
    # }
    # google = {
    #   source  = "hashicorp/google"
    #   version = ">= 5.30.0, < 6.0.0"
    # }
    # azurerm = {
    #   source  = "hashicorp/azurerm"
    #   version = ">= 3.100.0, < 4.0.0"
    # }
  }

  # Remote backend. Configure before any shared-environment `apply`.
  # Examples:
  #
  # backend "s3" {
  #   bucket         = "example-tfstate"
  #   key            = "envs/dev/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "example-tfstate-lock"
  #   encrypt        = true
  # }
  #
  # backend "gcs" {
  #   bucket = "example-tfstate"
  #   prefix = "envs/dev"
  # }
  #
  # backend "azurerm" {
  #   resource_group_name  = "tfstate-rg"
  #   storage_account_name = "exampletfstate"
  #   container_name       = "tfstate"
  #   key                  = "envs/dev/terraform.tfstate"
  # }
}
