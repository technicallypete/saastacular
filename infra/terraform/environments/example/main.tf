# The one committed root: one staging environment = one VM (gcp-staging-vm) plus its three DNS records
# (cloudflare-dns). All values come from a private terraform.tfvars and backend.hcl kept outside Git
# (see ../../README.md). GCP calls run as the impersonated deployer service account; the Cloudflare token comes
# from CLOUDFLARE_API_TOKEN in the operator environment. No secret value enters variables, state or outputs.

terraform {
  required_version = "~> 1.16"

  backend "gcs" {}

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "8.8.1"
    }

    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "5.28.1"
    }
  }
}

provider "google" {
  project                     = var.project_id
  region                      = var.region
  zone                        = var.zone
  impersonate_service_account = var.deployer_service_account_email
}

provider "cloudflare" {}

module "vm" {
  source = "../../modules/gcp-staging-vm"

  project_id               = var.project_id
  region                   = var.region
  zone                     = var.zone
  environment_name         = var.environment_name
  machine_type             = var.machine_type
  data_disk_size_gb        = var.data_disk_size_gb
  session_max_seconds      = var.session_max_seconds
  vm_service_account_email = var.vm_service_account_email
  operator_principals      = var.operator_principals
  labels                   = var.labels
}

module "dns" {
  source = "../../modules/cloudflare-dns"

  zone_id      = var.zone_id
  base_domain  = var.base_domain
  address      = module.vm.external_ip
  dns_ttl      = var.dns_ttl
  custom_label = var.custom_label
}
