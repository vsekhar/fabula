// TODO: use an existing network?

resource "google_compute_network" "network" {
    name = "service-${var.name}"
    auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnetwork" {
    name = "service-${var.name}"
    network = google_compute_network.network.self_link
    ip_cidr_range = "10.128.0.0/20"
    private_ip_google_access = true
}

resource "google_compute_firewall" "allow_ssh_from_iap" {
    lifecycle {
        create_before_destroy = true
    }
    name = "service-${var.name}-allow-ssh-from-iap"
    network = google_compute_network.network.name
    source_ranges = ["35.235.240.0/20"]
    direction = "INGRESS"
    allow {
        protocol = "tcp"
        ports =["22"]
    }
}

// Give internal hosts internet access
/*
resource "google_compute_router" "router" {
    name    = "service-test-router"
    network = google_compute_network.vpc.name
}

resource "google_compute_router_nat" "nat" {
    name = "service-test-nat"
    router = google_compute_router.router.name
    nat_ip_allocate_option             = "AUTO_ONLY"
    source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
*/
