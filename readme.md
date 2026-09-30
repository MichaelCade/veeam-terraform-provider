# Veeam Terraform Provider

A Personal project that will enable me to define as code my Veeam environment using Terraform and this provider. 

## Goals 

To be able to define various elements of my Veeam configuration via this terraform provider. 

- Infrastructure components - Proxies, Repositories 
- Backup Jobs - Across all workloads 
- Source Platform workload inventory 
- Credentials 

## Usage 

This provider has not been added to the Terraform Registry so have been testing locally first. To achieve this you need to use the `dev.tfrc` locally. 

```
export TF_CLI_CONFIG_FILE=/home/michael/Documents/veeam-terraform/examples/dev.tfrc
```

Under the examples folder there are a number of `.tf` files that can be used to create various objects within the Veeam setup. 

- `credentials.tf` - define credentials to be created 
- `hyperv_jobs.tf` - Not Tested, but will be used to define Hyper-V VM Backup Jobs 
- `infrastructure.tf` - Not Tested, but will be used to create Windows and Linux Proxies 
- `jobs.tf` - This is currently where we can define our VMware VM based backup jobs, This can be based on VMs or Tags
- `outputs.tf` - a data point created to output information, useful for when trying to find specific VMs or repositories
- `provider.tf` - Details on calling the local Veeam Terraform Provider and adding credential detail
- `repositories.tf` - Create all repository types, only tested with NFS and SMB targets
- `sobr.tf` - Not Tested, But will be used to create Scale out backup repositories with existing repositories 
- `unstructure_sources.tf` - To add Unstructured Data sources that require backups, only tested with SMB and NFS shares. 
- `variables.tf` - Holds variable info, used throughout the tf files. 
- `workload_jobs.tf` - Temp working location for new and additional backup job configuration 

I need to tidy some of this stuff up as I think we can make the above files clearer and more suited to each element we are wanting to deploy and configure. 

## Roadmap
Not in any order, but will depend on access to resources.

### Job Configuration 
- [ ] Test Proxmox Backup Job creation (workload_jobs.tf)
- [ ] Test Hyper-V Backup Job creation (workload_jobs.tf)
- [ ] Add vSphere Replication Jobs 
- [ ] Add Hyper-V Replication Jobs 
- [ ] Backup Copy Job Configuration
- [ ] SureBackup (This is a broad topic) (Application Groups, Virtual Labs, Jobs)
- [ ] Configuration Backup
- [ ] Advanced Job Settings 

### Credential Management
- [ ] Datacenter Credentials (We can create standard and SSH accounts)
- [ ] Cloud Credentials 
- [ ] Encryption Passwords 
- [ ] Key Management Servers 

### Target Repository 
- [ ] Test Object Storage Repository creation 
- [ ] Test Veeam Vault Repository creation
- [ ] Application Backup Repository 

### Public Cloud
- [ ] Add Veeam Backup for AWS 
- [ ] Add Veeam Backup for Google Cloud Platform (might need plugin installed)
- [ ] Add Veeam Backup for Microsoft Azure
- [ ] Configure Public Cloud Appliances 
- [ ] Create Public Cloud Policies 
- [ ] External Repositories

### Other Workload Sources

- [ ] File Server
- [ ] NAS Filer
- [ ] Object Storage (Source)
- [ ] EntraID
- [ ] Veeam Kasten
- [ ] Windows Agents
- [ ] Linux Agents
- [ ] MacOS Agents
- [ ] Enterprise Application Plug-ins (Oracle RMAN, MongoDB, Microsoft SQL Server and others)

### Hypervisors

- [ ] Add source platform vSphere
- [ ] Add source platform Proxmox
- [ ] Add source platform Hyper-V
- [ ] Red Hat OpenShift Virtualisation 
- [ ] Nutanix AHV 
- [ ] Red Hat Virtualisation (Not OpenShift)
- [ ] Oracle Linux Virtualisation Manager
- [ ] Scale Computing
- [ ] HPE Morpheus VM Essentials
- [ ] vCloud Director

### Other
- [ ] License Management
- [ ] Network Traffic Rules 
- [ ] Options 
- [ ] CDP Policy
- [ ] Tape
- [ ] Storage Integrations
- [ ] WAN Accelerators
- [ ] Is it possible to import an environment into code?
