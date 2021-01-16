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
        "roles/errorreporting.writer",
        "roles/monitoring.metricWriter",
        "roles/cloudtrace.agent",
        "roles/servicemanagement.serviceController", // for envoy
        "roles/compute.viewer", // for p2p discovery
    ])
    role    = each.value
    member  = "serviceAccount:${google_service_account.test.email}"
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

module "svc_id" {
    source = "../../../random_id"
    byte_length = 4
}

resource "google_endpoints_service" "grpc_service" {
    // services have a mandatory 30-day deletion process, so we need to
    // use a new name each time.
    service_name = "svctest-${module.svc_id.id}.endpoints.${var.project_id}.cloud.goog"
    grpc_config = templatefile("${path.module}/endpoints.tmpl.yaml",
        {
            service_name = "svctest-${module.svc_id.id}.endpoints.${var.project_id}.cloud.goog"
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

    versions = {
        "hello-v2" = {
            container_image = data.google_container_registry_image.hello-grpc
            machine_type = "e2-small"
            preemptible = true
            args = [
                "-port 8080",
                "-downstream ${module.internal_hello_service.service_name}:9500",
            ]
            // env = {}
            service_account = google_service_account.test.email
            envoy_config = {
                service_name = google_endpoints_service.grpc_service.service_name
                envoy_service_port = 8081
                backend_protocol = "grpc"
                backend_service_port = 8080
            }
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
    http_health_check_path = "/root"
    http_health_check_port = 8081
    service_to_container_ports = {
        "8081" = "8081" // health check --> envoy (8081) --> service (9500)
        "9500" = "9500"
    }

    versions = {
        "hello-v1" = {
            container_image = data.google_container_registry_image.hello-grpc
            machine_type = "e2-small"
            service_account = google_service_account.test.email
            preemptible = true
            args = [
                "-port 9500",
            ]
            // env = {}

            // To give somewhere for health checks to point to (upstream connects to
            // service directly).
            envoy_config = {
                service_name = google_endpoints_service.grpc_service.service_name
                envoy_service_port = 8081
                backend_protocol = "grpc"
                backend_service_port = 9500
            }
        }
    }
}
