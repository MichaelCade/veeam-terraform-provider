# Backup Repositories Example Configuration

# 1. Linux Local & Hardened Immutable Repository
# resource "veeam_repository_linux" "hardened_repo" {
#   name                      = "Linux Hardened Immutable Repo"
#   description               = "Immutable repository managed via Terraform"
#   host_id                   = "6745a759-2205-4cd2-b172-8ec8f7e60ef8" # GUID of Linux Managed Server
#   path                      = "/mnt/hardened_repository"
#   use_fast_cloning          = true
#   enable_governance_mode     = true
#   governance_retention_days = 30
# }

# 2. SMB NAS Share Repository
# resource "veeam_repository_smb" "nas_smb_repo" {
#   name           = "NAS SMB Share Repo"
#   description    = "SMB share backup target managed via Terraform"
#   share_path     = "\\\\readynas716\\data\\veeam_smb"
#   credentials_id = veeam_credential.smb_admin.id
# }

# 3. NFS NAS Share Repository
  resource "veeam_repository_nfs" "nas_nfs_repo" {
    name        = "NAS NFS Share Repo"
    description = "NFS share backup target managed via Terraform"
    share_path  = "192.168.169.3:/data/veeam_nfs"
  }

# 4. S3 Compatible Object Storage Repository (MinIO / Cloudian / Wasabi)
# resource "veeam_repository_s3_compatible" "minio_object_storage" {
#   name             = "MinIO S3 Object Storage Repo"
#   description      = "S3 compatible object storage managed via Terraform"
#   service_endpoint = "https://s3.lab.local:9000"
#   region_id        = "us-east-1"
#   bucket_name      = "veeam-backups"
#   folder_name      = "infrastructure-backups"
#   credentials_id   = veeam_credential.windows_admin.id
# }

# 5. Veeam Data Cloud Vault Repository
# resource "veeam_repository_veeam_vault" "cloud_vault" {
#   name                = "Veeam Data Cloud Vault Repo"
#   description         = "Veeam Data Cloud Vault storage managed via Terraform"
#   vault_id            = "00000000-0000-0000-0000-000000000000" # Vault account ID
#   folder              = "production-vault-backups"
#   enable_immutability = true
# }
