# NAS File Share Protection Jobs (NFS & SMB Share Targets)

# 1. NFS File Share Protection Job
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

# 2. SMB File Share Protection Job
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
