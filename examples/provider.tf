terraform {
  required_providers {
    veeam = {
      source  = "registry.terraform.io/michael/veeam"
      version = "0.1.0"
    }
  }
}

provider "veeam" {
  endpoint    = var.veeam_endpoint
  username    = var.veeam_username
  password    = var.veeam_password
  api_version = "1.3-rev2"
  insecure    = true
}
