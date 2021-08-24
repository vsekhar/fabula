module "regions" {
    source = "../regions"
    fabula_region = var.fabula_region
}

provider "google" {
    version = "~> 3.51.0"
    project = var.project_id
}

provider "google-beta" {
    version = "~> 3.51.0"
    project = var.project_id
}

resource "google_project_service" "service" {
    for_each = toset([
        "compute.googleapis.com",
        "container.googleapis.com",
        "containeranalysis.googleapis.com",
        "endpoints.googleapis.com",
        "iam.googleapis.com",
        "iamcredentials.googleapis.com", # for service account impersonation from k8s
        // "logging.googleapis.com", // TODO: enable to support direct logging from apps
        "pubsub.googleapis.com",
        "servicecontrol.googleapis.com", # for cloud endpoints
        "servicemanagement.googleapis.com", # for cloud endpoints
        "storage.googleapis.com"
    ])

    service = each.key
    disable_on_destroy = false
}

resource "google_service_account" "fabula" {
    account_id = "fabula-service-account"
    display_name = "Fabula service account"
}
