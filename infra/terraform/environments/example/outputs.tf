# Read by infra/staging (status, inventory, receipts). Names and addresses only.

output "instance_name" {
  value = module.vm.instance_name
}

output "instance_zone" {
  value = module.vm.instance_zone
}

output "instance_self_link" {
  value = module.vm.instance_self_link
}

output "external_ip" {
  value = module.vm.external_ip
}

output "service_account_email" {
  value = module.vm.service_account_email
}

output "data_disk_name" {
  value = module.vm.data_disk_name
}

output "network_name" {
  value = module.vm.network_name
}

output "subnetwork_name" {
  value = module.vm.subnetwork_name
}

output "firewall_names" {
  value = module.vm.firewall_names
}

output "max_run_duration_seconds" {
  value = module.vm.max_run_duration_seconds
}

output "dns_record_ids" {
  value = module.dns.record_ids
}
