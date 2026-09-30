variable "veeam_endpoint" {
  type        = string
  default     = "https://192.168.169.220:9419"
  description = "Veeam Backup & Replication REST API URL (Lab Linux Appliance)"
}

variable "veeam_username" {
  type        = string
  default     = "Administrator"
  description = "Veeam administrator username"
}

variable "veeam_password" {
  type        = string
  sensitive   = true
  description = "Veeam administrator password"
}

variable "agent_ssh_password" {
  type        = string
  sensitive   = true
  description = "Password for Linux agent/proxy credential resource"
}

variable "windows_admin_password" {
  type        = string
  sensitive   = true
  description = "Password for Windows administrator credential resource"
}

variable "smb_admin_password" {
  type        = string
  sensitive   = true
  description = "Password for SMB share access credential resource"
}

variable "hyperv_admin_password" {
  type        = string
  sensitive   = true
  description = "Password for Hyper-V administrator credential resource"
}


variable "linux_proxy_ip" {
  type        = string
  default     = "192.168.169.221"
  description = "IP address or DNS name of Linux Backup Proxy server to onboard"
}

variable "windows_repo_ip" {
  type        = string
  default     = "192.168.169.222"
  description = "IP address or DNS name of Windows Repository host to onboard"
}
