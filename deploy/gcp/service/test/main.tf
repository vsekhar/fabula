provider "google" {
    project = var.project_id
    region = "us-central1"
}

provider "google-beta" {
    project = var.project_id
    region = "us-central1"
}

data "google_container_registry_image" "hello-app-v1" {
    project = "google-samples"
    name = "hello-app:1.0"
}

data "google_container_registry_image" "hello-app-v2" {
    project = "google-samples"
    name = "hello-app:2.0"
}

module "test_group" {
    source = "../group"
    name = "test"
}

module "external_hello_service" {
    source = "../external"
    name = "external"
    group = module.test_group
    min_replicas = 1
    max_replicas = 2
    http_health_check_path = "/"
    http_health_check_port = 80
    service_to_container_ports = {
        "80" = "8080"
        "9619" = "9619"
    }

    versions = {
        "hello-v2" = {
            container_image = data.google_container_registry_image.hello-app-v2
            machine_type = "e2-small"
            preemptible = true
        }
    }
}

// To verify internal service:
//
//   1) SSH into an instance in external_hello_service
//   2) Curl internal_ip
//   3) verify version 1 server response.

module "internal_hello_service" {
    source = "../internal"
    name = "internal"
    group = module.test_group
    min_replicas = 1
    max_replicas = 2
    http_health_check_path = "/"
    http_health_check_port = 80
    service_to_container_ports = {
        "80" = "8080"
        "9619" = "9619"
    }

    versions = {
        "hello-v1" = {
            container_image = data.google_container_registry_image.hello-app-v1
            machine_type = "e2-small"
            preemptible = true
        }
    }
}
