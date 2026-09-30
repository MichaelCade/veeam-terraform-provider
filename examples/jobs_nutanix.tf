# Nutanix AHV Backup Job Example Configuration

# resource "veeam_job_nutanix" "ahv_cluster_backup" {
#   name          = "Nutanix AHV VM Backup Job"
#   description   = "Nutanix AHV workload backup managed via Terraform Provider"
#   repository_id = data.veeam_backup_repositories.all.repositories[0].id
#
#   retention_quantity = 7
#   retention_type     = "RestorePoints"
#
#   includes = [
#     {
#       name = "AHV-DB-01"
#       type = "VirtualMachine"
#     }
#   ]
# }
