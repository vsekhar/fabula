locals {
    target_tags = [
        "fabula-container-vm-mig"
    ]
}

data "google_compute_image" "cos" {
    family = "cos-stable"
    project = "cos-cloud"
}

resource "google_compute_instance_template" "fabula" {
    name_prefix = "fabula-instance-template-"
    project = var.project_id
    tags = local.target_tags
    labels = {
        container-vm = data.google_compute_image.cos.name
        container = data.google_container_registry_image.fabula.image_url
    }
    machine_type = "e2-small"
    disk {
        source_image = data.google_compute_image.cos.self_link
        auto_delete = true
        boot = true
    }
    metadata = {
        user-data = templatefile("${path.module}/gce_cloud-init.tmpl.yaml",
            {
                service_name = "fabula",
                // TODO: container_image_name = data.google_container_registry_image.fabula.image_url
                container_image_name = "gcr.io/google-samples/hello-app:2.0",
                host_to_container_ports = {
                    "80" = "8080",
                }
            }
        )
        google-logging-enabled = "true"
        google-monitoring-enabled = "true"
        cos-metrics-enabled = "true"
        enable-oslogin = "true"
    }

    network_interface {
        network = google_compute_network.vpc.self_link
        subnetwork = google_compute_subnetwork.subnet.self_link

        // Add this to provide direct access to instances via an ephemeral IP
        // access_config { network_tier = "PREMIUM" }
    }

    scheduling {
        preemptible         = "true"
        on_host_maintenance = "TERMINATE" // required for preemptible
        automatic_restart   = "false"     // required for preemptible
    }

    service_account {
        email = google_service_account.fabula.email
        scopes = [
            // Restrict via IAM on service account:
            // https://cloud.google.com/compute/docs/access/service-accounts#service_account_permissions
            "https://www.googleapis.com/auth/cloud-platform",
        ]
    }

    shielded_instance_config {
      enable_secure_boot = true
      enable_vtpm = true
      enable_integrity_monitoring = true
    }
}
