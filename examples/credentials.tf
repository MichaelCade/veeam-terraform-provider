# Credential Resources: Manage SSH & Windows Credentials in Veeam Credential Manager

resource "veeam_credential" "agent_ssh_key" {
  username    = "backup-agent-svc"
  password    = var.agent_ssh_password
  type        = "Linux"
  description = "Managed via Terraform for automated Linux workloads & proxies"
}

resource "veeam_credential" "windows_admin" {
  username    = "Administrator"
  password    = var.windows_admin_password
  type        = "Standard"
  description = "Managed via Terraform for Windows Hosts & Guest Processing"
}

resource "veeam_credential" "smb_admin" {
  username    = "terraform"
  password    = var.smb_admin_password
  type        = "Standard"
  description = "Managed via Terraform for Windows Hosts & Guest Processing"
}