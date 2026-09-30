# Veeam Terraform Provider (`terraform-provider-veeam`)

A Terraform provider that enables managing Veeam Backup & Replication (VBR) environments as code via the Veeam VBR REST API.

---

## Goals 

Define and manage core elements of your Veeam Backup & Replication infrastructure using Terraform:

- **Infrastructure Components**: Managed Servers (Linux Proxies & Windows Repositories), Backup Repositories (NFS, SMB, Hardened Linux, Windows, S3 Compatible).
- **Backup Jobs**: VMware vSphere VM/Tag backup jobs, NAS & File Share protection jobs, Hyper-V, Proxmox VE, and Nutanix AHV backup jobs.
- **Data Sources**: Live inventory browsing (VMware vSphere VMs & Tags), Managed Servers, Repositories, Jobs, and Unstructured Data Sources.
- **Credentials**: Standard Windows & Linux SSH credentials in Veeam Credential Manager.

---

## Local Development & Usage 

To test the provider locally before publishing to the Terraform Registry, use `dev.tfrc`:

```bash
export TF_CLI_CONFIG_FILE=/home/michael/Documents/veeam-terraform/examples/dev.tfrc
```

### Example Configurations (`examples/`)

- `provider.tf` - Declares local provider override and connection details.
- `variables.tf` - Input variable definitions (`veeam_endpoint`, `veeam_username`, `veeam_password`, etc.).
- `terraform.tfvars.example` - Template file for copying to `terraform.tfvars` (ignored by Git).
- `credentials.tf` - Manages SSH & Windows credentials (`veeam_credential`).
- `infrastructure.tf` - Manages Linux Proxies & Windows Repositories (`veeam_managed_server_linux`, `veeam_managed_server_windows`).
- `repositories.tf` - Manages backup repository targets (`veeam_repository_nfs`, `veeam_repository_smb`, `veeam_repository_linux`, `veeam_repository_s3_compatible`).
- `sobr.tf` - Scale-Out Backup Repository (SOBR) resources & data sources (`veeam_sobr_repository`).
- `unstructured_sources.tf` - Registers SMB & NFS file share backup sources (`veeam_unstructured_data_smb_share`, `veeam_unstructured_data_nfs_share`).
- `jobs_vmware.tf` - VMware vSphere VM & Tag backup jobs (`veeam_job_vmware`).
- `jobs_nas.tf` - NAS File Share protection jobs (`veeam_job_file_share`).
- `jobs_hyperv.tf` - Microsoft Hyper-V backup job examples (`veeam_job_hyperv`).
- `jobs_proxmox.tf` - Proxmox VE backup job examples (`veeam_job_proxmox`).
- `jobs_nutanix.tf` - Nutanix AHV backup job examples (`veeam_job_nutanix`).
- `outputs.tf` - Outputs discovered inventory items, job GUIDs, and repository metadata.

---

## Importing Existing Infrastructure

All provider resources support `terraform import` to adopt existing Veeam infrastructure into code:

```bash
# Import an existing VMware Backup Job by GUID
terraform import 'veeam_job_vmware.sql_vm_backup[0]' <JOB_GUID>

# Import an existing NFS Share Source by GUID
terraform import veeam_unstructured_data_nfs_share.nas_nfs_source <SOURCE_GUID>

# Import an existing Credential entry by GUID
terraform import veeam_credential.windows_admin <CREDENTIAL_GUID>
```

---

## Roadmap

### Job Configuration 
- [x] VMware vSphere VM & Tag Backup Jobs (`veeam_job_vmware`)
- [x] NAS & File Share Protection Jobs (`veeam_job_file_share`)
- [ ] Test Proxmox Backup Job creation (`veeam_job_proxmox`)
- [ ] Test Hyper-V Backup Job creation (`veeam_job_hyperv`)
- [ ] Add vSphere Replication Jobs 
- [ ] Add Hyper-V Replication Jobs 
- [ ] Backup Copy Job Configuration
- [ ] SureBackup (Application Groups, Virtual Labs, Jobs)
- [ ] Configuration Backup
- [ ] Advanced Job Settings 

### Credential Management
- [x] Datacenter Credentials (Standard Windows & Linux SSH accounts via `veeam_credential`)
- [x] Sensitive Password Decoupling & `.gitignore` Security Rules
- [ ] Cloud Credentials 
- [ ] Encryption Passwords 
- [ ] Key Management Servers 

### Target Repository 
- [x] NFS NAS Share Repository (`veeam_repository_nfs`)
- [x] SMB NAS Share Repository (`veeam_repository_smb`)
- [x] Linux Local & Hardened Immutable Repository (`veeam_repository_linux`)
- [x] Windows Local Repository (`veeam_repository_windows`)
- [x] Scale-Out Backup Repository (`veeam_sobr_repository`)
- [ ] Test Object Storage Repository creation (`veeam_repository_s3_compatible`)
- [ ] Test Veeam Vault Repository creation (`veeam_repository_veeam_vault`)
- [ ] Application Backup Repository 

### Public Cloud
- [ ] Add Veeam Backup for AWS 
- [ ] Add Veeam Backup for Google Cloud Platform
- [ ] Add Veeam Backup for Microsoft Azure
- [ ] Configure Public Cloud Appliances 
- [ ] Create Public Cloud Policies 
- [ ] External Repositories

### Other Workload Sources
- [x] NAS Filer & SMB/NFS File Share Sources (`veeam_unstructured_data_smb_share`, `veeam_unstructured_data_nfs_share`)
- [ ] File Server
- [ ] Object Storage (Source)
- [ ] EntraID
- [ ] Veeam Kasten
- [ ] Windows Agents
- [ ] Linux Agents
- [ ] MacOS Agents
- [ ] Enterprise Application Plug-ins (Oracle RMAN, MongoDB, Microsoft SQL Server)

### Hypervisors
- [x] VMware vSphere / vCenter Server Registration (`veeam_managed_server_vsphere`) & Inventory Browsing (`veeam_inventory`)
- [x] Microsoft Hyper-V Host & Cluster Registration (`veeam_managed_server_hyperv`) & Backup Jobs (`veeam_job_hyperv`)
- [x] Proxmox VE Node & Cluster Registration (`veeam_managed_server_proxmox`) & Backup Jobs (`veeam_job_proxmox`)
- [ ] Red Hat OpenShift Virtualisation 
- [ ] Nutanix AHV 
- [ ] Red Hat Virtualisation (Not OpenShift)
- [ ] Oracle Linux Virtualisation Manager
- [ ] Scale Computing
- [ ] HPE Morpheus VM Essentials
- [ ] vCloud Director

### Other
- [x] Is it possible to import an environment into code? (`terraform import` implemented for all resources)
- [ ] License Management
- [ ] Network Traffic Rules 
- [ ] Options 
- [ ] CDP Policy
- [ ] Tape
- [ ] Storage Integrations
- [ ] WAN Accelerators
