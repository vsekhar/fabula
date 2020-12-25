data "google_compute_image" "cos" {
    family = "cos-stable"
    project = "cos-cloud"
}

resource "random_id" "tmpl" {
  byte_length = 4
}

resource "google_compute_instance_template" "template" {
    lifecycle {
        create_before_destroy = true
    }

    name = "${var.name}-${random_id.tmpl.id}" // name_prefix creates very long names
    tags = var.tags
    labels = {
        // labels must be [a-z0-9_-] and at most 63 characters
        container-vm = data.google_compute_image.cos.name
        container-image-project = var.container_image.project
        container-image-name = replace(var.container_image.name, "/[:.]/", "-")
        container-image-digest = (
            var.container_image.digest != null
            ? substr(replace(var.container_image.digest, "/[:]/", "-"), 0, 15)
            : null
        )
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
                service_name = var.name,
                container_image_name = var.container_image.image_url
                host_to_container_ports = var.host_to_container_ports
                args = var.args
            }
        )
        google-logging-enabled = "true"
        google-monitoring-enabled = "true"
        cos-metrics-enabled = "true"
        enable-oslogin = "true"
    }

    network_interface {
        network = (var.network != "" ? var.network : null)
        subnetwork = (var.subnetwork != "" ? var.subnetwork : null)

        dynamic "access_config" {
            for_each = var.public_ip == "" ? [] : [1]
            content {
                network_tier = "PREMIUM"
                nat_ip = (var.public_ip != "0.0.0.0" ? var.public_ip : null)
            }
        }
    }

    scheduling {
        preemptible         = var.preemptible
        on_host_maintenance = "TERMINATE" // required for preemptible
        automatic_restart   = "false"     // required for preemptible
    }

    service_account {
        email = var.service_account
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
