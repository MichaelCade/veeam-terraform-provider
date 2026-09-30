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


