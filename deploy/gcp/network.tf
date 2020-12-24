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

resource "google_compute_firewall" "fabula-allow-external" {
  lifecycle {
    create_before_destroy = true
  }
  name = "fabula-allow-external"
  network = google_compute_network.vpc.name
  allow {
    protocol = "icmp" // ping
  }
  allow {
    protocol = "tcp"
    ports = ["80", "8080"]
  }
}

resource "google_compute_firewall" "fabula-allow-internal" {
  lifecycle {
    create_before_destroy = true
  }
  name = "fabula-allow-internal"
  network = google_compute_network.vpc.name
  source_ranges = [ "10.128.0.0/9" ]
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
}

resource "google_compute_firewall" "fabula_allow_ssh_from_iap" {
    lifecycle {
        create_before_destroy = true
    }
    name = "fabula-allow-ssh-from-iap"
    network = google_compute_network.vpc.name
    source_ranges = ["35.235.240.0/20"]
    direction = "INGRESS"
    allow {
        protocol = "tcp"
        ports =["22"]
    }
}
