# One non-Spot VM with a reserved static address and a separate retained data disk. desired_status is
# deliberately unset: the lifecycle commands stop and start the VM with gcloud and Terraform does not manage the
# power state. The startup script goes in the plain metadata key (updated in place), never in a recreating field.
# The session length (max_run_duration) is set from session_max_seconds when the VM is created; an existing VM's length
# is written by infra/staging/staging.sh start --duration (gcloud compute instances set-scheduling while TERMINATED,
# then a refresh-only apply), because hashicorp/google 8.8.1 replaces the VM for that field (G070 A2). Terraform
# therefore ignores it after creation: -var session_max_seconds no longer changes an existing VM.

locals {
  edge_tag         = "${var.environment_name}-edge"
  iap_tag          = "${var.environment_name}-iap"
  iap_source_range = "35.235.240.0/20"
  data_device_name = "staging-data"
  labels           = merge(var.labels, { environment = var.environment_name, "managed-by" = "terraform" })
  network_tags     = var.iap_ssh ? [local.edge_tag, local.iap_tag] : [local.edge_tag]
  startup_script   = join("\n", ["#!/bin/bash", "export STAGING_ENV=${var.environment_name}", file("${path.module}/startup.sh")])
}

resource "google_compute_address" "ip" {
  project      = var.project_id
  name         = "${var.environment_name}-ip"
  region       = var.region
  address_type = "EXTERNAL"
  network_tier = "PREMIUM"
  labels       = local.labels
}

# Images, containers and the named volumes (database, object store, edge certificates) live here; startup.sh
# formats it once (filesystem label staging-data) and mounts it at /var/lib/docker before Docker starts.
resource "google_compute_disk" "data" {
  project = var.project_id
  name    = "${var.environment_name}-data"
  zone    = var.zone
  type    = var.disk_type
  size    = var.data_disk_size_gb
  labels  = local.labels
}

resource "google_compute_instance" "vm" {
  project      = var.project_id
  name         = "${var.environment_name}-vm"
  zone         = var.zone
  machine_type = var.machine_type
  tags         = local.network_tags
  labels       = local.labels

  metadata = {
    enable-oslogin = "TRUE"
    startup-script = local.startup_script
  }

  boot_disk {
    auto_delete = true

    initialize_params {
      image  = var.boot_image
      size   = var.boot_disk_size_gb
      type   = var.disk_type
      labels = local.labels
    }
  }

  # The device name is the /dev/disk/by-id/google-staging-data path startup.sh waits for.
  attached_disk {
    source      = google_compute_disk.data.id
    device_name = local.data_device_name
    mode        = "READ_WRITE"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.id

    access_config {
      nat_ip       = google_compute_address.ip.address
      network_tier = "PREMIUM"
    }
  }

  service_account {
    email  = var.vm_service_account_email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  scheduling {
    provisioning_model          = "STANDARD"
    preemptible                 = false
    automatic_restart           = true
    on_host_maintenance         = "MIGRATE"
    instance_termination_action = var.instance_termination_action

    max_run_duration {
      seconds = var.session_max_seconds
    }
  }

  # Only the session length, nothing broader (see the header); the refresh-only apply records the live value.
  lifecycle {
    ignore_changes = [scheduling[0].max_run_duration]
  }
}
