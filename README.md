# Cpanel Terraform Provider

The Cpanel Terraform Provider allows you to manage Cpanel accounts and their associated resources using Terraform.

## Installation

To use this provider, you need to install it. You can do this by adding it to your Terraform configuration.

Terraform Configuration
Add the following to your main.tf file to use the Jumpserver provider:

```hcl
terraform {
  required_providers {
    jumpserver = {
      source  = "bhfonseca/cpanel"
      version = "~> 0.0.2-rc1"
    }
  }
}

provider "cpanel" {
  host      = "cpanel.example.com"
  username  = "your_user"
  api_token = "your_token"
  insecure  = true
}
```

## Resources

This provider supports the following resources:

* `cpanel_zone_record`: Manage DNS zones in Cpanel.

## Resources Definitions

* [DNS Zone](./docs/resources/cpanel_zone_record.md): Manage DNS zones in Cpanel.