# Names and addresses only; no output carries credential material.

output "instance_name" {
  description = "VM name (<env>-vm)."
  value       = google_compute_instance.vm.name
}

output "instance_zone" {
  description = "VM zone."
  value       = google_compute_instance.vm.zone
}

output "instance_self_link" {
  description = "VM self link."
  value       = google_compute_instance.vm.self_link
}

output "external_ip" {
  description = "Reserved static external IPv4 address of the VM."
  value       = google_compute_address.ip.address
}

output "service_account_email" {
  description = "Attached VM runtime service account."
  value       = var.vm_service_account_email
}

output "data_disk_name" {
  description = "Retained data disk name (<env>-data)."
  value       = google_compute_disk.data.name
}

output "network_name" {
  description = "Dedicated VPC name (<env>-vpc)."
  value       = google_compute_network.vpc.name
}

output "subnetwork_name" {
  description = "Dedicated subnet name (<env>-subnet)."
  value       = google_compute_subnetwork.subnet.name
}

output "firewall_names" {
  description = "Ingress rule names (<env>-edge and, with iap_ssh, <env>-iap)."
  value       = concat([google_compute_firewall.edge.name], google_compute_firewall.iap[*].name)
}

output "max_run_duration_seconds" {
  description = "Native session limit in seconds as the state records it (set at creation; start --duration refreshes it)."
  value       = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds
}
