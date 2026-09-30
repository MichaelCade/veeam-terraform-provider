# Scale-Out Backup Repository (SOBR) Example Configuration

# Query all existing Scale-Out Backup Repositories
data "veeam_scale_out_repositories" "all" {}

# Create a Scale-Out Backup Repository (SOBR) combining Performance Extents
# resource "veeam_sobr_repository" "global_sobr" {
#   name        = "Enterprise Scale-Out Repository"
#   description = "Managed via Terraform Provider"
#   performance_extent_ids = [
#     data.veeam_backup_repositories.all.repositories[0].id,
#     data.veeam_backup_repositories.all.repositories[1].id,
#   ]
# }

output "scale_out_repositories" {
  value       = data.veeam_scale_out_repositories.all.repositories
  description = "List of configured Scale-Out Backup Repositories"
}
