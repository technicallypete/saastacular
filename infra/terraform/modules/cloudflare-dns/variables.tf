# Inputs of the three staging address records. The Cloudflare token is provider configuration taken from
# CLOUDFLARE_API_TOKEN in the operator environment; it is never a variable.

variable "zone_id" {
  description = "Cloudflare zone ID (32 hexadecimal characters) of the zone that holds base_domain."
  type        = string

  validation {
    condition     = can(regex("^[0-9a-f]{32}$", var.zone_id))
    error_message = "zone_id must be a Cloudflare zone ID (32 lower-case hexadecimal characters)."
  }
}

variable "base_domain" {
  description = "Staging base domain B; records B, *.B and <custom_label>.B are managed."
  type        = string

  validation {
    condition     = can(regex("^([a-z0-9]([a-z0-9-]*[a-z0-9])?\\.)+[a-z]{2,}$", var.base_domain))
    error_message = "base_domain must be a lower-case DNS name."
  }
}

variable "address" {
  description = "IPv4 address the records point to (the VM's reserved static address)."
  type        = string

  validation {
    condition     = can(cidrnetmask("${var.address}/32"))
    error_message = "address must be an IPv4 address."
  }
}

variable "dns_ttl" {
  description = "Record TTL in seconds (1 = automatic)."
  type        = number
  default     = 300

  validation {
    condition     = var.dns_ttl == 1 || (var.dns_ttl >= 60 && var.dns_ttl <= 86400)
    error_message = "dns_ttl must be 1 (automatic) or 60-86400."
  }
}

variable "custom_label" {
  description = "Label of the custom-domain CNAME target record (<custom_label>.B)."
  type        = string
  default     = "custom"

  validation {
    condition     = can(regex("^[a-z0-9]([a-z0-9-]*[a-z0-9])?$", var.custom_label))
    error_message = "custom_label must be one lower-case DNS label."
  }
}
