# Infrastructure Resources & Data Sources

# Managed Servers: Onboard Linux Backup Proxies & Windows Repository Hosts
# Note: Set var.linux_proxy_ip / var.windows_repo_ip to valid reachable IP/DNS in your environment
# resource "veeam_managed_server_linux" "proxy_host" {
#   name           = var.linux_proxy_ip
#   description    = "Linux Backup Proxy Host managed via Terraform"
#   credentials_id = veeam_credential.agent_ssh_key.id
#   ssh_port       = 22
# }
#
# resource "veeam_managed_server_windows" "repo_host" {
#   name           = var.windows_repo_ip
#   description    = "Windows Repository Host managed via Terraform"
#   credentials_id = veeam_credential.windows_admin.id
# }

# Hypervisor Infrastructure Servers (vSphere / vCenter, Microsoft Hyper-V)
# resource "veeam_managed_server_vsphere" "vcenter_server" {
#   name           = "192.168.169.181"
#   description    = "vCenter Server managed via Terraform Provider"
#   credentials_id = veeam_credential.windows_admin.id
#   port           = 443
# }
#
# resource "veeam_managed_server_hyperv" "hyperv_standalone_host" {
#   name           = "192.168.169.180"
#   description    = "Hyper-V Host managed via Terraform Provider"
#   server_type    = "HvServer" # Or "HvCluster"
#   credentials_id = veeam_credential.hyperv_admin.id
# }

# Data Sources: Read current repositories and managed servers
data "veeam_backup_repositories" "all" {}
data "veeam_managed_servers" "all" {}
