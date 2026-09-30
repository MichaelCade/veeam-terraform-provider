# Unstructured Data Backup Sources Example (SMB & NFS File Shares)

# 1. SMB Share Source (registered in Veeam Inventory)
 resource "veeam_unstructured_data_smb_share" "nas_smb_source" {
   path           = "\\\\readynas716.vzilla.local\\data\\90daysofdevops"
   credentials_id = veeam_credential.smb_admin.id
 }

# 2. NFS Share Export Source (registered in Veeam Inventory)
resource "veeam_unstructured_data_nfs_share" "nas_nfs_source" {
  path = "readynas716.vzilla.local:/data/veeam-source-nfs"
}

# 3. Data Source to list all registered Unstructured Data Sources in Veeam
data "veeam_unstructured_data_servers" "all_sources" {}

output "unstructured_data_sources" {
  value       = data.veeam_unstructured_data_servers.all_sources.servers
  description = "List of all registered unstructured data sources (SMB, NFS, Object Storage)."
}

output "nfs_source_id" {
  value       = veeam_unstructured_data_nfs_share.nas_nfs_source.id
  description = "GUID of the NFS share source in Veeam Inventory."
}
