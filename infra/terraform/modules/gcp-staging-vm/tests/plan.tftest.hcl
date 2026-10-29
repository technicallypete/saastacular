# Offline assertions for the gcp-staging-vm module (G070 Plan P0.3 as rewritten by r1.2). The google provider is
# mocked: no credentials, no network, nothing allocated. Run from the module directory:
#   terraform init -backend=false && terraform test

mock_provider "google" {}

variables {
  project_id               = "example-project"
  region                   = "example-region1"
  zone                     = "example-region1-a"
  environment_name         = "staging-test"
  vm_service_account_email = "staging-test-vm@example-project.iam.gserviceaccount.com"
  operator_principals      = ["user:ops@example.net"]
}

run "defaults" {
  command = plan

  assert {
    condition     = length(google_compute_firewall.edge.allow) == 1 && one(google_compute_firewall.edge.allow).protocol == "tcp"
    error_message = "The edge rule must allow exactly one tcp entry."
  }

  assert {
    condition     = toset(one(google_compute_firewall.edge.allow).ports) == toset(["80", "443"])
    error_message = "The edge rule must allow tcp/80 and tcp/443 only."
  }

  assert {
    condition     = toset(google_compute_firewall.edge.source_ranges) == toset(["0.0.0.0/0"]) && google_compute_firewall.edge.direction == "INGRESS"
    error_message = "The edge rule must be ingress from 0.0.0.0/0."
  }

  assert {
    condition     = length(google_compute_firewall.edge.deny) == 0 && toset(google_compute_firewall.edge.target_tags) == toset(["staging-test-edge"])
    error_message = "The edge rule must target the edge tag and carry no deny entry."
  }

  assert {
    condition     = length(google_compute_firewall.iap) == 1 && length(google_compute_firewall.iap[0].allow) == 1 && one(google_compute_firewall.iap[0].allow).protocol == "tcp"
    error_message = "The IAP rule must allow exactly one tcp entry."
  }

  assert {
    condition     = toset(one(google_compute_firewall.iap[0].allow).ports) == toset(["22"]) && toset(google_compute_firewall.iap[0].source_ranges) == toset(["35.235.240.0/20"])
    error_message = "SSH must be allowed only from the IAP range 35.235.240.0/20."
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 7200
    error_message = "The default native session limit must be 7200 seconds (2 h)."
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].instance_termination_action == "STOP"
    error_message = "Expiry must stop the VM."
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].provisioning_model == "STANDARD" && google_compute_instance.vm.scheduling[0].preemptible == false
    error_message = "The VM must be a standard (non-Spot, non-preemptible) VM."
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].automatic_restart == true && google_compute_instance.vm.scheduling[0].on_host_maintenance == "MIGRATE"
    error_message = "Host maintenance must live-migrate and restart the VM."
  }

  assert {
    condition     = google_compute_instance.vm.desired_status == null
    error_message = "Terraform must not manage the power state."
  }

  assert {
    condition     = toset(keys(google_compute_instance.vm.metadata)) == toset(["enable-oslogin", "startup-script"])
    error_message = "Instance metadata keys must be exactly enable-oslogin and startup-script."
  }

  assert {
    condition     = google_compute_instance.vm.metadata["enable-oslogin"] == "TRUE"
    error_message = "OS Login must be enabled."
  }

  assert {
    condition     = startswith(google_compute_instance.vm.metadata["startup-script"], "#!/bin/bash\nexport STAGING_ENV=staging-test\n#!/usr/bin/env bash\n")
    error_message = "The startup script must be startup.sh behind the environment-name line."
  }

  assert {
    condition     = google_compute_instance.vm.labels["environment"] == "staging-test" && google_compute_instance.vm.labels["managed-by"] == "terraform"
    error_message = "The VM must carry the environment and managed-by labels."
  }

  assert {
    condition     = google_compute_disk.data.labels["environment"] == "staging-test" && google_compute_disk.data.labels["managed-by"] == "terraform"
    error_message = "The data disk must carry the environment and managed-by labels."
  }

  assert {
    condition     = google_compute_address.ip.labels["environment"] == "staging-test" && google_compute_address.ip.labels["managed-by"] == "terraform"
    error_message = "The static address must carry the environment and managed-by labels."
  }

  assert {
    condition     = google_compute_instance.vm.boot_disk[0].initialize_params[0].labels["managed-by"] == "terraform"
    error_message = "The boot disk must carry the managed-by label."
  }

  assert {
    condition     = google_compute_instance.vm.name == "staging-test-vm" && google_compute_disk.data.name == "staging-test-data" && google_compute_address.ip.name == "staging-test-ip"
    error_message = "VM, data disk and address names must be deterministic."
  }

  assert {
    condition     = google_compute_network.vpc.name == "staging-test-vpc" && google_compute_network.vpc.auto_create_subnetworks == false && google_compute_subnetwork.subnet.name == "staging-test-subnet"
    error_message = "The VM must use a dedicated custom-mode VPC and subnet."
  }

  assert {
    condition     = google_compute_firewall.edge.name == "staging-test-edge" && google_compute_firewall.iap[0].name == "staging-test-iap"
    error_message = "Firewall rule names must be deterministic."
  }

  assert {
    condition     = google_compute_instance.vm.machine_type == "e2-standard-2" && google_compute_instance.vm.boot_disk[0].initialize_params[0].image == "debian-cloud/debian-12"
    error_message = "Default machine type and boot image changed."
  }

  assert {
    condition     = google_compute_instance.vm.boot_disk[0].initialize_params[0].size == 20 && google_compute_disk.data.size == 50 && google_compute_disk.data.type == "pd-balanced"
    error_message = "Default disk sizes or type changed."
  }

  assert {
    condition     = length(google_compute_instance.vm.attached_disk) == 1 && google_compute_instance.vm.attached_disk[0].device_name == "staging-data"
    error_message = "The data disk must be attached as device staging-data (startup.sh contract)."
  }

  assert {
    condition     = google_compute_instance.vm.service_account[0].email == "staging-test-vm@example-project.iam.gserviceaccount.com" && contains(google_compute_instance.vm.service_account[0].scopes, "https://www.googleapis.com/auth/cloud-platform")
    error_message = "The VM must run as the bootstrap-created service account with the cloud-platform scope."
  }

  assert {
    condition     = google_compute_instance.vm.shielded_instance_config[0].enable_secure_boot && google_compute_instance.vm.shielded_instance_config[0].enable_vtpm && google_compute_instance.vm.shielded_instance_config[0].enable_integrity_monitoring
    error_message = "The VM must be a Shielded VM."
  }

  assert {
    condition     = length(google_compute_instance.vm.network_interface) == 1 && length(google_compute_instance.vm.network_interface[0].access_config) == 1
    error_message = "The VM must have one interface with one external access config."
  }

  assert {
    condition     = toset(google_compute_instance.vm.tags) == toset(["staging-test-edge", "staging-test-iap"])
    error_message = "The VM must carry exactly the edge and IAP tags."
  }

  assert {
    condition     = length(google_compute_instance_iam_member.os_login) == 1 && google_compute_instance_iam_member.os_login["user:ops@example.net"].role == "roles/compute.osAdminLogin"
    error_message = "Listed operators get OS Login administrator access on this VM only."
  }

  assert {
    condition     = output.max_run_duration_seconds == 7200
    error_message = "The max_run_duration_seconds output must report the configured limit."
  }
}

# No output name may look like credential material.
run "outputs_carry_no_credential_names" {
  command = plan

  assert {
    condition     = length([for name in keys(run.defaults) : name if can(regex("(?i)secret|token|password|key", name))]) == 0
    error_message = "An output name matches secret|token|password|key."
  }

  assert {
    condition     = length(keys(run.defaults)) == 10
    error_message = "The module output set changed; review it for credential material and update this count."
  }
}

run "session_length_is_configurable" {
  command = plan

  variables {
    session_max_seconds = 600
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 600 && output.max_run_duration_seconds == 600
    error_message = "session_max_seconds must set max_run_duration (600 s, the lower bound, is accepted)."
  }
}

run "session_length_upper_bound" {
  command = plan

  variables {
    session_max_seconds = 28800
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 28800 && output.max_run_duration_seconds == 28800
    error_message = "28800 s (8 h), the upper bound, must be accepted and set max_run_duration."
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].instance_termination_action == "STOP"
    error_message = "Expiry must stop the VM at every session length."
  }
}

run "without_iap_ssh" {
  command = plan

  variables {
    iap_ssh             = false
    operator_principals = []
  }

  assert {
    condition     = length(google_compute_firewall.iap) == 0 && toset(google_compute_instance.vm.tags) == toset(["staging-test-edge"])
    error_message = "iap_ssh = false must leave no SSH rule and no IAP tag."
  }

  assert {
    condition     = length(google_compute_instance_iam_member.os_login) == 0
    error_message = "No operator binding may exist without operator_principals."
  }
}

run "labels_cannot_override_fixed_keys" {
  command = plan

  variables {
    labels = { environment = "other", "managed-by" = "hand", team = "ops" }
  }

  assert {
    condition     = google_compute_instance.vm.labels["environment"] == "staging-test" && google_compute_instance.vm.labels["managed-by"] == "terraform" && google_compute_instance.vm.labels["team"] == "ops"
    error_message = "Extra labels must merge without replacing environment or managed-by."
  }
}

# Any termination action other than STOP is refused, including the one that would remove the VM.
run "rejects_other_termination_action" {
  command = plan

  variables {
    instance_termination_action = "TERMINATE"
  }

  expect_failures = [var.instance_termination_action]
}

run "rejects_session_below_minimum" {
  command = plan

  variables {
    session_max_seconds = 599
  }

  expect_failures = [var.session_max_seconds]
}

run "rejects_session_above_maximum" {
  command = plan

  variables {
    session_max_seconds = 28801
  }

  expect_failures = [var.session_max_seconds]
}

run "rejects_fractional_session" {
  command = plan

  variables {
    session_max_seconds = 1.5
  }

  expect_failures = [var.session_max_seconds]
}

# Inside the range but not whole seconds: only the whole-number condition refuses it.
run "rejects_fractional_session_in_range" {
  command = plan

  variables {
    session_max_seconds = 1800.5
  }

  expect_failures = [var.session_max_seconds]
}

run "rejects_invalid_environment_name" {
  command = plan

  variables {
    environment_name = "Staging_Env"
  }

  expect_failures = [var.environment_name]
}

run "rejects_zone_outside_region" {
  command = plan

  variables {
    zone = "other-region2-a"
  }

  expect_failures = [var.zone]
}

run "rejects_default_compute_service_account" {
  command = plan

  variables {
    vm_service_account_email = "123456789-compute@developer.gserviceaccount.com"
  }

  expect_failures = [var.vm_service_account_email]
}

# An existing VM's session length is not Terraform's to change (instance.tf lifecycle.ignore_changes; the provider
# would replace the VM): after an apply creates the VM, a plan with another session_max_seconds keeps the recorded
# length, which proves ignore_changes is effective. The self link assertions are only a sanity check: the mock provider
# does not model the real provider's RequiresReplace on max_run_duration, so they do not prove that no replacement is
# planned. That proof is the live refresh-only plan (the L-5 scope check refuses any replacement).
# Applied runs come last, so every plan run above starts from an empty state; the 600 s chain has its own state.
run "create_vm_7200" {
  command = apply

  variables {
    session_max_seconds = 7200
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 7200 && output.max_run_duration_seconds == 7200
    error_message = "The created VM must carry session_max_seconds (7200)."
  }
}

run "existing_vm_7200_ignores_600" {
  command = plan

  variables {
    session_max_seconds = 600
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 7200 && output.max_run_duration_seconds == 7200
    error_message = "An existing VM's max_run_duration must stay 7200 when session_max_seconds is 600 (ignore_changes)."
  }

  assert {
    condition     = output.instance_self_link == run.create_vm_7200.instance_self_link
    error_message = "The instance self link must stay the same (sanity check; the mock does not model RequiresReplace)."
  }
}

run "create_vm_600" {
  command   = apply
  state_key = "session-600"

  variables {
    session_max_seconds = 600
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 600 && output.max_run_duration_seconds == 600
    error_message = "The created VM must carry session_max_seconds (600)."
  }
}

run "existing_vm_600_ignores_7200" {
  command   = plan
  state_key = "session-600"

  variables {
    session_max_seconds = 7200
  }

  assert {
    condition     = google_compute_instance.vm.scheduling[0].max_run_duration[0].seconds == 600 && output.max_run_duration_seconds == 600
    error_message = "An existing VM's max_run_duration must stay 600 when session_max_seconds is 7200 (ignore_changes)."
  }

  assert {
    condition     = output.instance_self_link == run.create_vm_600.instance_self_link
    error_message = "The instance self link must stay the same (sanity check; the mock does not model RequiresReplace)."
  }
}
