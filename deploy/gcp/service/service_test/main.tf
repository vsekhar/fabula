provider "google" {
    project = var.project_id
    region = "us-central1"
}

provider "google-beta" {
    project = var.project_id
    region = "us-central1"
}

data "google_container_registry_image" "hello-app" {
    project = "google-samples"
    name = "hello-app:2.0"
}

resource "google_compute_network" "vpc" {
    name = "service-test"
}

module "hello-template" {
    source = "../../container_vm"
    name = "hello-app"
    container_image = data.google_container_registry_image.hello-app
    host_to_container_ports = {
        "80" = "8080"
    }
    preemptible = true
    machine_type = "e2-small"
    network = google_compute_network.vpc.name
}

resource "google_compute_region_health_check" "hc" {
    name = "service-test-hc"
    check_interval_sec = 2
    timeout_sec = 2
    http_health_check {
      port = "80"
      request_path = "/"
    }
}

resource "google_compute_health_check" "hc" {
    name = "http-health-check-rigm"
    check_interval_sec = 2
    timeout_sec = 2
    http_health_check {
      port = "80"
      request_path = "/"
    }
}

resource "google_compute_firewall" "allow_health_checks" {
    name = "service-test-allow-health-checks"
    network = google_compute_network.vpc.name

    // https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges
    source_ranges = [ "35.191.0.0/16", "130.211.0.0/22" ]
    direction = "INGRESS"
    allow {
        protocol = "tcp"
        ports = ["80"]
    }
}

resource "google_compute_firewall" "allow-external" {
    name = "service-test-allow-external"
    network = google_compute_network.vpc.name
    allow {
        protocol = "icmp" // ping
    }
    allow {
        protocol = "tcp"
        ports = ["80", "8080"]
}
}

module "hello_service" {
    source = "../"
    name = "test"
    region = "us-central1"
    min_replicas = 1
    global_health_check = google_compute_health_check.hc.id
    region_health_checks = [google_compute_region_health_check.hc.id]

    versions = [
        {
            name = "hello-app"
            instance_template = module.hello-template.self_link
        }
    ]
}
