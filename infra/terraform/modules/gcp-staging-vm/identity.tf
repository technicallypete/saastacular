# The VM runtime and deployer service accounts, their Secret Manager, Artifact Registry and state-bucket bindings
# are created once by bootstrap/bootstrap.sh; this module only attaches the VM account (instance.tf) and may grant
# OS Login administrator access on this one VM to listed operators. IAP tunnel access stays a project grant.

resource "google_compute_instance_iam_member" "os_login" {
  for_each = toset(var.operator_principals)

  project       = var.project_id
  zone          = var.zone
  instance_name = google_compute_instance.vm.name
  role          = "roles/compute.osAdminLogin"
  member        = each.value
}
