# Veeam Terraform Provider - Presentation Guide & Talk Track

This guide provides a slide-by-slide script, timing recommendations, technical deep-dives, and Q&A strategies for presenting the **Veeam Terraform Provider** (`terraform-provider-veeam`).

---

## ⏱️ Recommended Deck Timing (Total: 25 - 30 minutes)

- **Slides 1–3: Introduction & Core Value** (5 mins)
- **Slides 4–7: Multi-Hypervisor & Storage Technical Capabilities** (10 mins)
- **Slides 8–10: GitOps Workflow & Code Walkthrough** (8 mins)
- **Slides 11–12: Roadmap & Q&A Call to Action** (5 mins)

---

## 🎤 Slide-by-Slide Speaker Script & Technical Context

### Slide 1: Title & Introduction
- **Talk Track**: *"Welcome everyone. Today we are introducing `terraform-provider-veeam`—a Terraform provider created to bring native Infrastructure as Code to Veeam Backup & Replication. Until now, enterprise backup infrastructure and backup policies were often managed manually via GUI click-ops or ad-hoc scripts. This provider bridges the gap between modern cloud automation and data protection, allowing engineers to declare backup targets, managed hypervisors, and backup jobs directly in HCL alongside their application workloads."*
- **Key Visual Focus**: Point out the primary pillars: Multi-Hypervisor Backup, Linux Immutability, REST API v12 native integration, and zero-downtime brownfield import.

---

### Slide 2: The Paradigm Shift (Manual Backup Ops vs GitOps)
- **Talk Track**: *"Why bring backup management into code? In traditional operations, backup configuration drift is a silent killer. A backup repository folder or job schedule modified by hand during an outage often stays undocumented. When disaster strikes, rebuilding your backup ecosystem manually takes hours or days. With Terraform, your Git commit log becomes the single source of truth for your entire data protection posture."*
- **Key Takeaways**:
  - **Auditability**: Every change passes through code review (Git PRs).
  - **Speed**: Deploying new sites or DR environments takes minutes via `terraform apply`.

---

### Slide 3: Architecture & REST API Integration
- **Talk Track**: *"Architecturally, the provider acts as a translator between Terraform Core and the Veeam VBR REST API v12+. Built with the modern Go-based Terraform Plugin Framework, it handles bearer token session management, credential encryption decoupling, schema validation, and lifecycle state management automatically."*
- **Technical Detail**: Emphasize that no custom agents or direct DB modifications are required; all calls strictly adhere to Veeam's supported REST API schemas.

---

### Slide 4: Multi-Hypervisor Workload Support
- **Talk Track**: *"Veeam has always been celebrated for its multi-hypervisor versatility. The Terraform provider brings native resources for every major platform: VMware vSphere, Proxmox VE, Microsoft Hyper-V, and Nutanix AHV."*
- **Code References**:
  - `veeam_managed_server_vsphere` & `veeam_job_vmware`
  - `veeam_managed_server_proxmox` & `veeam_job_proxmox`
  - `veeam_managed_server_hyperv` & `veeam_job_hyperv`

---

### Slide 5: Storage Targets & Ransomware Immutability
- **Talk Track**: *"Storage target provisioning is a highlight of the provider. In ransomware defense, Hardened Linux Repositories with immutability flags are the gold standard. Using `veeam_repository_linux`, you can set `immutability_days = 30` directly in HCL. The provider also supports Scale-Out Backup Repositories (SOBR), S3 Compatible Object Storage, and SMB/NFS targets."*
- **Security Note**: Mention sensitive credential decoupling—SSH keys and passwords are managed securely in the Veeam Credential Vault via `veeam_credential`.

---

### Slide 6: Unstructured Data & NAS Protection
- **Talk Track**: *"Veeam Backup & Replication protects unstructured NAS data across corporate file filers. With `veeam_unstructured_data_smb_share` and `veeam_unstructured_data_nfs_share`, engineers can onboard file share paths as backup sources and create dedicated file protection jobs (`veeam_job_file_share`) to enforce retention and archive policies."*

---

### Slide 7: Zero Downtime Brownfield Adoption (`terraform import`)
- **Talk Track**: *"A common question from backup admins is: 'Do I have to tear down my existing VBR setup to use Terraform?' The answer is a clear NO. Every single resource in `terraform-provider-veeam` supports standard `terraform import`. You simply find the GUID of an existing job, repository, or credential and run `terraform import`. Terraform adopts the object into state without touching the running backup schedules."*

---

### Slide 8: GitOps & CI/CD Pipeline Integration
- **Talk Track**: *"Here is where the real power shines in a DevOps pipeline. When a developer or cloud architect provisions a new cluster or set of application VMs in Terraform, the same pipeline can declare the Veeam protection policy. If a VM is instantiated, it is immediately protected. No un-backed orphan infrastructure ever reaches production."*

---

### Slide 9: Declarative Code Showcase
- **Talk Track**: *"Let's look at a concrete HCL configuration. In under 30 lines of code, we register a Proxmox host (`veeam_managed_server_proxmox`), provision a Hardened Linux Immutable Backup Repository (`veeam_repository_linux`), and tie them together into an active Proxmox backup job (`veeam_job_proxmox`). The dependency graph is automatically computed by Terraform."*

---

### Slide 10: Live Data Inventory & Discovery
- **Talk Track**: *"You don't need to hardcode GUIDs in your Terraform code. Data sources like `veeam_inventory`, `veeam_backup_repositories`, and `veeam_managed_servers` allow you to dynamically discover existing vSphere tags, VMs, or backup repositories at runtime and feed them into your job definitions."*

---

### Slide 11: Project Roadmap & Expansion
- **Talk Track**: *"The project is moving fast. Core hypervisors, repository targets, NAS protection, and import capabilities are fully functional today. Looking ahead, our roadmap includes Veeam Backup for Public Cloud (AWS, Azure, GCP), Veeam Kasten K10 Kubernetes backup integration, Enterprise Application plug-ins (Oracle, SAP, SQL), and CDP policies."*

---

### Slide 12: Conclusion & Quickstart Call to Action
- **Talk Track**: *"To summarize: `terraform-provider-veeam` empowers your team to automate backup infrastructure, prevent configuration drift, and guarantee cyber resilience. You can start testing locally today using the dev override configuration (`examples/dev.tfrc`) provided in the repository. Thank you, and let's open the floor to Q&A!"*

---

## ❓ Antipatterns & Frequently Asked Questions (Q&A Prep)

### Q1: What version of Veeam Backup & Replication is required?
**Answer**: Veeam Backup & Replication **v12 or higher** is required, as the provider relies on the expanded VBR REST API endpoints.

### Q2: Does `terraform destroy` delete actual backup data?
**Answer**: Resource destruction via `terraform destroy` removes the job definition, repository registration, or managed server entry from Veeam Backup & Replication configuration. Actual backup files on disk or object storage remain intact according to Veeam safety defaults.

### Q3: How are sensitive passwords and SSH keys stored?
**Answer**: Sensitive credentials are created or imported into Veeam Credential Manager (`veeam_credential`) and referenced in other resources via GUIDs or credential IDs. Passwords do not need to be stored in plain text in version control.

### Q4: Can I test this provider locally before it is published to the Terraform Registry?
**Answer**: Yes! Set `export TF_CLI_CONFIG_FILE=$PWD/examples/dev.tfrc`. This instructs the Terraform CLI to use your locally compiled provider binary directly.
