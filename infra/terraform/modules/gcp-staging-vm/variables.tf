# Inputs of one staging VM. No variable carries a secret: provider credentials come from the operator's ADC
# (service-account impersonation) and application secrets live only in Secret Manager.

variable "project_id" {
  description = "GCP project that holds this staging environment."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", var.project_id))
    error_message = "project_id must be a GCP project ID."
  }
}

variable "region" {
  description = "Region of the subnet and the static address."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]+[0-9]$", var.region))
    error_message = "region must be a GCP region name."
  }
}

variable "zone" {
  description = "Zone of the VM and its data disk; must lie in region."
  type        = string

  validation {
    condition     = startswith(var.zone, "${var.region}-")
    error_message = "zone must be a zone of region."
  }
}

variable "environment_name" {
  description = "Environment name; prefixes every resource name and is the environment label value."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,20}$", var.environment_name))
    error_message = "environment_name must match ^[a-z][a-z0-9-]{1,20}$."
  }
}

variable "machine_type" {
  description = "Compute Engine machine type (changing it needs a stopped VM)."
  type        = string
  default     = "e2-standard-2"
}

variable "boot_image" {
  description = "Boot image as project/family; the startup script supports Debian 12 only."
  type        = string
  default     = "debian-cloud/debian-12"
}

variable "boot_disk_size_gb" {
  description = "Boot disk size in GB (operating system and Docker packages only)."
  type        = number
  default     = 20

  validation {
    condition     = var.boot_disk_size_gb >= 10
    error_message = "boot_disk_size_gb must be at least 10."
  }
}

variable "data_disk_size_gb" {
  description = "Retained data disk size in GB, mounted at /var/lib/docker (images, containers, named volumes)."
  type        = number
  default     = 50

  validation {
    condition     = var.data_disk_size_gb >= 10
    error_message = "data_disk_size_gb must be at least 10."
  }
}

variable "disk_type" {
  description = "Persistent disk type of the boot and data disks."
  type        = string
  default     = "pd-balanced"

  validation {
    condition     = contains(["pd-balanced", "pd-ssd", "pd-standard"], var.disk_type)
    error_message = "disk_type must be pd-balanced, pd-ssd or pd-standard."
  }
}

variable "subnet_cidr" {
  description = "Primary range of the dedicated subnet."
  type        = string
  default     = "10.20.0.0/24"

  validation {
    condition     = can(cidrhost(var.subnet_cidr, 0))
    error_message = "subnet_cidr must be an IPv4 CIDR range."
  }
}

variable "session_max_seconds" {
  description = "Native session limit (max_run_duration) set when the VM is created; Compute Engine stops the VM when it elapses and re-arms it at every start. An existing VM ignores it (infra/staging/staging.sh start --duration changes the length with gcloud set-scheduling and a refresh-only apply). Policy range 600 to 28800 (10 min to 8 h)."
  type        = number
  default     = 7200

  validation {
    condition     = var.session_max_seconds >= 600 && var.session_max_seconds <= 28800 && floor(var.session_max_seconds) == var.session_max_seconds
    error_message = "session_max_seconds must be a whole number from 600 to 28800 (10 min to 8 h)."
  }
}

variable "instance_termination_action" {
  description = "Action when the session limit elapses. Only STOP is accepted: expiry never removes the VM or its disks."
  type        = string
  default     = "STOP"

  validation {
    condition     = var.instance_termination_action == "STOP"
    error_message = "instance_termination_action must be STOP."
  }
}

variable "iap_ssh" {
  description = "Allow tcp/22 from the IAP TCP-forwarding range only (OS Login over IAP); no other SSH source exists."
  type        = bool
  default     = true
}

variable "vm_service_account_email" {
  description = "The bootstrap-created VM runtime service account (never the default compute account)."
  type        = string

  validation {
    condition     = endswith(var.vm_service_account_email, ".iam.gserviceaccount.com")
    error_message = "vm_service_account_email must be a user-managed service account (*.iam.gserviceaccount.com)."
  }
}

variable "operator_principals" {
  description = "Additional members (user:, group: or serviceAccount:) granted OS Login administrator access on this VM only."
  type        = list(string)
  default     = []

  validation {
    condition     = alltrue([for p in var.operator_principals : can(regex("^(user|group|serviceAccount):\\S+@\\S+$", p))])
    error_message = "operator_principals entries must be user:, group: or serviceAccount: members."
  }
}

variable "labels" {
  description = "Extra labels; environment and managed-by are always set by the module and cannot be overridden."
  type        = map(string)
  default     = {}
}
