provider "google" {
    project = var.project_id
    region = "us-central1"
}

provider "google-beta" {
    project = var.project_id
    region = "us-central1"
}

resource "google_project_service" "service" {
    for_each = toset([
        "compute.googleapis.com",
        "endpoints.googleapis.com",
        "iam.googleapis.com",
        "pubsub.googleapis.com",
        "servicecontrol.googleapis.com", # for cloud endpoints
        "servicemanagement.googleapis.com", # for cloud endpoints
    ])

    service = each.key
    disable_on_destroy = false
}

resource "google_service_account" "test" {
    account_id = "svc-test"
    display_name = "Test service account"
}

resource "google_project_iam_member" "iam" {
    for_each = toset([
        "roles/logging.logWriter",
        "roles/monitoring.metricWriter",
        "roles/cloudtrace.agent",
        "roles/servicemanagement.serviceController",
    ])
    role    = each.value
    member  = "serviceAccount:${google_service_account.test.email}"
}

data "google_container_registry_image" "hello-app-v1" {
    project = "google-samples"
    name = "hello-app:1.0"
}

data "google_container_registry_image" "hello-app-v2" {
    project = "google-samples"
    name = "hello-app:2.0"
}

data "google_container_registry_image" "hello-grpc" {
    project = "fabula-resources"
    name = "hello"
    tag = "latest"
}

module "test_group" {
    source = "../group"
    name = "test"
}

resource "google_endpoints_service" "grpc_service" {
    service_name = "svctest.endpoints.${var.project_id}.cloud.goog"
    grpc_config = templatefile("${path.module}/endpoints.tmpl.yaml",
        {
            project_id = var.project_id
        }
    )
    protoc_output_base64 = filebase64("${path.module}/api_descriptor.pb")
}

resource "google_project_service" "endpoint_service" {
    service = google_endpoints_service.grpc_service.service_name
    // disable_on_destroy = false
}

module "external_hello_service" {
    source = "../external"
    name = "external"
    group = module.test_group
    min_replicas = 1
    max_replicas = 2
    http_health_check_path = "/root"
    http_health_check_port = 80
    service_to_container_ports = {
        "80" = "8081" // http(80) --> envoy(8081) --> server(8080)
    }
    envoy_config = {
        // TODO: make envoy config per-version
        service_name = google_endpoints_service.grpc_service.service_name
        // TODO: envoy_to_container_ports
        envoy_service_port = 8081
        backend_protocol = "grpc"
        backend_service_port = 8080 // internal container port (process listen port)
    }
    service_account = google_service_account.test.email

    versions = {
        "hello-v2" = {
            container_image = data.google_container_registry_image.hello-grpc
            machine_type = "e2-small"
            preemptible = true
            args = [
                "-port 8080",
            ]
            // env = {}
        }
    }
}

// To verify internal service:
//
//   1) SSH into an instance in external_hello_service
//   2) Curl internal_service_name
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
    service_account = google_service_account.test.email

    versions = {
        "hello-v1" = {
            container_image = data.google_container_registry_image.hello-app-v1
            machine_type = "e2-small"
            preemptible = true
            // args = []
            // env = {}
        }
    }
}
