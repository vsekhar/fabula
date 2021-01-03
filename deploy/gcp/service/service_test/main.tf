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

module "hello_service" {
    source = "../"
    name = "test"
    region = "us-central1"
    min_replicas = 1
    network = google_compute_network.vpc.name
    http_health_check_path = "/"
    http_health_check_port = 80
    // TODO: internal and external
    service_to_container_ports = {
        "80" = "8080"
    }

    versions = {
        "hello-app" = {
            container_image = data.google_container_registry_image.hello-app
            machine_type = "e2-small"
            preemptible = true
        }
    }
}
