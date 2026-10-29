# Offline assertions for the cloudflare-dns module (G070 Plan P0.3 as rewritten by r1.2). The cloudflare provider
# is mocked: no token, no network, no record written. Run from the module directory:
#   terraform init -backend=false && terraform test

mock_provider "cloudflare" {}

variables {
  zone_id     = "0123456789abcdef0123456789abcdef"
  base_domain = "staging.example.net"
  address     = "192.0.2.10"
}

run "three_dns_only_a_records" {
  command = plan

  assert {
    condition     = length(cloudflare_dns_record.this) == 3
    error_message = "Exactly three records must be managed."
  }

  assert {
    condition     = toset([for r in cloudflare_dns_record.this : r.name]) == toset(["staging.example.net", "*.staging.example.net", "custom.staging.example.net"])
    error_message = "The records must be B, *.B and custom.B."
  }

  assert {
    condition     = alltrue([for r in cloudflare_dns_record.this : r.type == "A"])
    error_message = "Every record must be an A record (no TXT, CNAME or AAAA)."
  }

  assert {
    condition     = alltrue([for r in cloudflare_dns_record.this : r.proxied == false])
    error_message = "Every record must be DNS-only (proxied = false)."
  }

  assert {
    condition     = alltrue([for r in cloudflare_dns_record.this : r.content == "192.0.2.10" && r.ttl == 300 && r.zone_id == "0123456789abcdef0123456789abcdef"])
    error_message = "Every record must point at the address in the given zone with the default TTL 300."
  }

  assert {
    condition     = alltrue([for r in cloudflare_dns_record.this : r.comment == "managed-by=terraform"])
    error_message = "Every record must carry the managed-by comment."
  }
}

# No output name may look like credential material.
run "outputs_carry_no_credential_names" {
  command = plan

  assert {
    condition     = length([for name in keys(run.three_dns_only_a_records) : name if can(regex("(?i)secret|token|password|key", name))]) == 0
    error_message = "An output name matches secret|token|password|key."
  }

  assert {
    condition     = length(keys(run.three_dns_only_a_records)) == 1
    error_message = "The module output set changed; review it for credential material and update this count."
  }
}

run "ttl_and_custom_label_are_configurable" {
  command = plan

  variables {
    dns_ttl      = 60
    custom_label = "domains"
  }

  assert {
    condition     = contains([for r in cloudflare_dns_record.this : r.name], "domains.staging.example.net") && alltrue([for r in cloudflare_dns_record.this : r.ttl == 60])
    error_message = "custom_label and dns_ttl must be applied."
  }
}

run "rejects_invalid_zone_id" {
  command = plan

  variables {
    zone_id = "replace-me"
  }

  expect_failures = [var.zone_id]
}

run "rejects_ttl_below_minimum" {
  command = plan

  variables {
    dns_ttl = 30
  }

  expect_failures = [var.dns_ttl]
}

run "rejects_upper_case_base_domain" {
  command = plan

  variables {
    base_domain = "Staging.Example.NET"
  }

  expect_failures = [var.base_domain]
}

run "rejects_non_ipv4_address" {
  command = plan

  variables {
    address = "2001:db8::10"
  }

  expect_failures = [var.address]
}
