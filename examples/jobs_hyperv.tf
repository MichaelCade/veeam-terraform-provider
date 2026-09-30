# Microsoft Hyper-V Backup Jobs Example Configuration

# resource "veeam_job_hyperv" "hyperv_cluster_backup" {
#   name          = "Hyper-V Production VM Backup Job"
#   description   = "Hyper-V workload backup managed via Terraform Provider"
#   repository_id = data.veeam_backup_repositories.all.repositories[0].id
#
#   retention_quantity = 7
#   retention_type     = "RestorePoints"
#
#   includes = [
#     {
#       name      = "hv-app-01"
#       type      = "VirtualMachine"
#       host_name = "hyperv-cluster-01.lab.local"
#     }
#   ]
# }
