provider "template" {
    version = "~> 2.2.0"
}

resource "google_project_iam_binding" "fabula_esp_agent" {
    role = "roles/cloudtrace.agent"
    members = [
        "serviceAccount:${google_service_account.fabula.email}",
    ]
}

resource "google_project_iam_binding" "fabula_esp_service_controller" {
    role = "roles/servicemanagement.serviceController"
    members = [
        "serviceAccount:${google_service_account.fabula.email}",
    ]
}

resource "google_endpoints_service" "grpc_service" {
    service_name = "fabula.endpoints.${var.project_id}.cloud.goog"
    grpc_config = templatefile("${path.module}/endpoints.tmpl.yaml",
        {
            project_id = var.project_id
        }
    )
    protoc_output_base64 = filebase64("${path.module}/../api_descriptor.pb")
}

resource "google_project_service" "fabula_endpoint_service" {
    service = google_endpoints_service.grpc_service.service_name
    disable_on_destroy = false
}
