# VMware vSphere Backup Jobs Management

data "veeam_jobs" "all" {}

data "veeam_inventory" "sql_server" {
  host_name      = "192.168.169.181"
  hierarchy_type = "VmsAndTemplates"
  name_filter    = "SQLServer-1"
}

data "veeam_inventory" "linux_tag" {
  host_name      = "192.168.169.181"
  hierarchy_type = "VmsAndTags"
  name_filter    = "Linux Desktop Backup"
}

# 1. Create Backup Job for VM "SQLServer-1" if found
resource "veeam_job_vmware" "sql_vm_backup" {
  count         = length(data.veeam_inventory.sql_server.items) > 0 ? 1 : 0
  name          = "SQLServer-1 Backup Job"
  description   = "vSphere VM SQLServer-1 VM Backup Job managed via Terraform Provider"
  repository_id = data.veeam_backup_repositories.all.repositories[0].id

  includes = [
    {
      name      = data.veeam_inventory.sql_server.items[0].name
      type      = data.veeam_inventory.sql_server.items[0].type
      host_name = "192.168.169.181"
      object_id = data.veeam_inventory.sql_server.items[0].object_id
      urn       = data.veeam_inventory.sql_server.items[0].urn
    }
  ]

  schedule {
    run_automatically = true
    daily {
      is_enabled = true
      local_time = "23:00"
      daily_kind = "Everyday"
    }
  }

  gfs_policy {
    is_enabled = true
    weekly {
      is_enabled     = true
      keep_for_weeks = 4
    }
    monthly {
      is_enabled      = true
      keep_for_months = 12
    }
  }

  guest_processing {
    app_aware_processing_enabled = true
    guest_indexing_enabled       = true
    guest_credentials_id         = veeam_credential.windows_admin.id
  }
}

# 2. Create Backup Job for vSphere Tag "Linux Desktop Backup" if found
resource "veeam_job_vmware" "linux_tag_backup" {
  count         = length(data.veeam_inventory.linux_tag.items) > 0 ? 1 : 0
  name          = "Linux Desktop Tag Backup Job"
  description   = "vSphere Linux Desktop Backup tag managed via Terraform Provider"
  repository_id = data.veeam_backup_repositories.all.repositories[0].id

  includes = [
    {
      name      = data.veeam_inventory.linux_tag.items[0].name
      type      = data.veeam_inventory.linux_tag.items[0].type
      host_name = "192.168.169.181"
      object_id = data.veeam_inventory.linux_tag.items[0].object_id
      urn       = data.veeam_inventory.linux_tag.items[0].urn
    }
  ]
}
