# B, *.B and <custom_label>.B as DNS-only A records. Proxying stays off: the edge terminates TLS itself and the
# wildcard needs no Cloudflare certificate. No AAAA record. _acme-challenge TXT records belong to the edge (DNS-01)
# and any tenant CNAME/verification record to the domain-qualification owner; Terraform manages neither.

locals {
  records = {
    base     = var.base_domain
    wildcard = "*.${var.base_domain}"
    custom   = "${var.custom_label}.${var.base_domain}"
  }
}

resource "cloudflare_dns_record" "this" {
  for_each = local.records

  zone_id = var.zone_id
  name    = each.value
  type    = "A"
  content = var.address
  proxied = false
  ttl     = var.dns_ttl
  comment = "managed-by=terraform"
}
