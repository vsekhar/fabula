resource "google_compute_network" "vpc" {
  name                    = "${var.project_id}-vpc"
  auto_create_subnetworks = "false"
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.project_id}-subnet"
  region        = module.regions.gce_region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.10.0.0/24"

}

resource "google_compute_firewall" "fabula" {
  name = "fabula-firewall"
  network = google_compute_network.vpc.name
  allow {
    protocol = "icmp" // ping
  }
  allow {
    protocol = "tcp"
    ports = ["80"]
  }
}
