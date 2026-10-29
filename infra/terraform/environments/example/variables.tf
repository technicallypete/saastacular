# Root inputs; the modules validate them. Values live in the private terraform.tfvars (see
# terraform.tfvars.example). None of them is a secret.

variable "project_id" {
  description = "GCP project of this staging environment."
  type        = string
}

variable "region" {
  description = "GCP region."
  type        = string
}

variable "zone" {
  description = "GCP zone in region."
  type        = string
}

variable "environment_name" {
  description = "Environment name (^[a-z][a-z0-9-]{1,20}$); resource-name prefix and environment label."
  type        = string
}

variable "deployer_service_account_email" {
  description = "Bootstrap-created deployer service account that Terraform impersonates."
  type        = string
}

variable "vm_service_account_email" {
  description = "Bootstrap-created VM runtime service account."
  type        = string
}

variable "operator_principals" {
  description = "Extra members granted OS Login administrator access on the VM."
  type        = list(string)
  default     = []
}

variable "machine_type" {
  description = "Compute Engine machine type."
  type        = string
  default     = "e2-standard-2"
}

variable "data_disk_size_gb" {
  description = "Retained data disk size in GB."
  type        = number
  default     = 50
}

variable "session_max_seconds" {
  description = "Native session limit in seconds (600 to 28800) at VM creation; infra/staging/staging.sh passes it on every plan, but an existing VM ignores it (start --duration sets the length with gcloud set-scheduling and a refresh-only apply)."
  type        = number
  default     = 7200
}

variable "labels" {
  description = "Extra labels for labelled resources."
  type        = map(string)
  default     = {}
}

variable "zone_id" {
  description = "Cloudflare zone ID of the zone holding base_domain (not the GCP zone)."
  type        = string
}

variable "base_domain" {
  description = "Staging base domain B."
  type        = string
}

variable "dns_ttl" {
  description = "DNS record TTL in seconds."
  type        = number
  default     = 300
}

variable "custom_label" {
  description = "Label of the custom-domain CNAME target record."
  type        = string
  default     = "custom"
}
