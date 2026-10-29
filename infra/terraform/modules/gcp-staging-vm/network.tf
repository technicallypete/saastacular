# Dedicated VPC: the implied ingress deny applies to everything except the two rules below; egress stays open
# (apt, ACME, Cloudflare API, Artifact Registry, Secret Manager).

resource "google_compute_network" "vpc" {
  project                 = var.project_id
  name                    = "${var.environment_name}-vpc"
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"
}

resource "google_compute_subnetwork" "subnet" {
  project                  = var.project_id
  name                     = "${var.environment_name}-subnet"
  region                   = var.region
  network                  = google_compute_network.vpc.id
  ip_cidr_range            = var.subnet_cidr
  private_ip_google_access = true
}

# Public HTTP/HTTPS to the edge only; Compose publishes nothing else.
resource "google_compute_firewall" "edge" {
  project       = var.project_id
  name          = "${var.environment_name}-edge"
  network       = google_compute_network.vpc.id
  description   = "Public tcp/80 and tcp/443 to the staging edge."
  direction     = "INGRESS"
  source_ranges = ["0.0.0.0/0"]
  target_tags   = [local.edge_tag]

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }
}

# SSH only from Google's IAP TCP-forwarding range; there is no external SSH path.
resource "google_compute_firewall" "iap" {
  count = var.iap_ssh ? 1 : 0

  project       = var.project_id
  name          = "${var.environment_name}-iap"
  network       = google_compute_network.vpc.id
  description   = "tcp/22 from the IAP TCP-forwarding range only."
  direction     = "INGRESS"
  source_ranges = [local.iap_source_range]
  target_tags   = [local.iap_tag]

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}
