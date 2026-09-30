output "found_sql_vm" {
  value       = data.veeam_inventory.sql_server.items
  description = "Discovered inventory details for SQLServer-1"
}

output "found_linux_tag" {
  value       = data.veeam_inventory.linux_tag.items
  description = "Discovered inventory details for Linux Desktop Backup tag"
}

output "sql_job_id" {
  value       = length(veeam_job_vmware.sql_vm_backup) > 0 ? veeam_job_vmware.sql_vm_backup[0].id : "Not Created"
  description = "GUID of the SQLServer-1 backup job"
}

output "all_repositories" {
  value       = data.veeam_backup_repositories.all.repositories
  description = "List of all backup repositories currently in Veeam Backup & Replication"
}
