# Infrastructure Resources & Data Sources

# Managed Servers: Onboard Linux Backup Proxies & Windows Repository Hosts
# Note: Set var.linux_proxy_ip / var.windows_repo_ip to valid reachable IP/DNS in your environment
#resource "veeam_managed_server_linux" "proxy_host" {
#  name           = var.linux_proxy_ip
#  description    = "Linux Backup Proxy Host managed via Terraform"
#  credentials_id = veeam_credential.agent_ssh_key.id
#  ssh_port       = 22
#}
#
#resource "veeam_managed_server_windows" "repo_host" {
#  name           = var.windows_repo_ip
#  description    = "Windows Repository Host managed via Terraform"
#  credentials_id = veeam_credential.windows_admin.id
#}

# Data Sources: Read current repositories and managed servers
data "veeam_backup_repositories" "all" {}
data "veeam_managed_servers" "all" {}
