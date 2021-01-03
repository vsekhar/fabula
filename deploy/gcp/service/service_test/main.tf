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

    versions = [
        {
            name = "hello-app"
            instance_template = module.hello-template.self_link
        }
    ]
}
