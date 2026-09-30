# Proxmox VE Backup Job Example Configuration

# resource "veeam_job_proxmox" "pve_cluster_backup" {
#   name          = "Proxmox Production VM Backup Job"
#   description   = "Proxmox VE workload backup managed via Terraform Provider"
#   repository_id = data.veeam_backup_repositories.all.repositories[0].id
#
#   retention_quantity = 14
#   retention_type     = "RestorePoints"
#
#   includes = [
#     {
#       name = "pve-app-server-01"
#       type = "VirtualMachine"
#     }
#   ]
# }
