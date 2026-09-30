# Multi-Hypervisor & Workload Backup Jobs Example Configuration

# 1. Proxmox VE Backup Job
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

# 2. Nutanix AHV Backup Job
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

# 3. NFS File Share Backup Job
resource "veeam_job_file_share" "nfs_file_backup" {
  name               = "NFS File Share Protection Job"
  description        = "NFS share backup job managed via Terraform Provider"
  repository_id      = data.veeam_backup_repositories.all.repositories[0].id
  retention_quantity = 7
  retention_type     = "Days"

  objects = [
    {
      file_server_id = veeam_unstructured_data_nfs_share.nas_nfs_source.id
      path           = veeam_unstructured_data_nfs_share.nas_nfs_source.path
    }
  ]

  schedule {
    run_automatically = true
    daily {
      is_enabled = true
      local_time = "23:30"
      daily_kind = "Everyday"
    }
  }
}

# 4. SMB File Share Backup Job
resource "veeam_job_file_share" "smb_file_backup" {
  name               = "SMB File Share Protection Job"
  description        = "SMB share backup job managed via Terraform Provider"
  repository_id      = data.veeam_backup_repositories.all.repositories[0].id
  retention_quantity = 7
  retention_type     = "Days"

  objects = [
    {
      file_server_id = veeam_unstructured_data_smb_share.nas_smb_source.id
      path           = veeam_unstructured_data_smb_share.nas_smb_source.path
    }
  ]

  schedule {
    run_automatically = true
    daily {
      is_enabled = true
      local_time = "23:45"
      daily_kind = "Everyday"
    }
  }
}
