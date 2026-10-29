output "record_ids" {
  description = "Cloudflare record IDs by role (base, wildcard, custom), for the inventory and absence checks."
  value       = { for role, record in cloudflare_dns_record.this : role => record.id }
}
