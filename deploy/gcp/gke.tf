resource "google_compute_network" "vpc" {
  name                    = "${var.project_id}-vpc"
  auto_create_subnetworks = "false"
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.project_id}-subnet"
  region        = module.regions.gce_region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.10.0.0/24"

}

resource "google_container_cluster" "fabula" {
    provider = google-beta
    name = "${var.project_id}-gke"
    location = module.regions.gce_region

    remove_default_node_pool = true
    initial_node_count = 1

    network    = google_compute_network.vpc.name
    subnetwork = google_compute_subnetwork.subnet.name
    networking_mode = "VPC_NATIVE"
    ip_allocation_policy {}

    master_auth {
        username = ""
        password = ""

        client_certificate_config {
            issue_client_certificate = false
        }
    }

    workload_identity_config {
        identity_namespace = "${var.project_id}.svc.id.goog"
    }
}

resource "google_container_node_pool" "fabula_preemptible_nodes" {
    name = "${google_container_cluster.fabula.name}-node-pool"
    location = module.regions.gce_region
    cluster = google_container_cluster.fabula.name
    node_count = 1
    lifecycle {
        ignore_changes = [
            initial_node_count
        ]
    }

    autoscaling {
        min_node_count = 1 // per region (3 total)
        max_node_count = 4 // 12 total
    }

    management {
        auto_repair = true
        auto_upgrade = true
    }

    node_config {
        preemptible = true
        machine_type = "f1-micro"
        tags = [
            "gke-node",
            "${var.project_id}-gke",
        ]
        metadata = {
            disable-legacy-endpoints = "true"
        }
    }
}

data "google_client_config" "provider" {}

provider "kubernetes" {
    alias = "fabula_k8s_provider"
    version = "~> 1.13.3"
    load_config_file = "false"

    host  = "https://${google_container_cluster.fabula.endpoint}"
    token = data.google_client_config.provider.access_token
    cluster_ca_certificate = base64decode(
        google_container_cluster.fabula.master_auth[0].cluster_ca_certificate,
    )
}

module "fabula_kubernetes" {
    source = "../k8s"
    providers = {
        kubernetes: kubernetes.fabula_k8s_provider
    }

    service_account_email = google_service_account.fabula.email
    storage_bucket_name = google_storage_bucket.pack_storage.name
    endpoints_service_name = google_endpoints_service.grpc_service.service_name
    fabula_image = data.google_container_registry_image.fabula.image_url
}

resource "google_service_account_iam_binding" "workload_identity" {
    service_account_id = google_service_account.fabula.name
    role = "roles/iam.workloadIdentityUser"
    members = [
        "serviceAccount:${var.project_id}.svc.id.goog[${module.fabula_kubernetes.kubernetes_service_account_id}]"
    ]
}
