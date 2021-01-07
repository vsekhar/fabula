data "google_compute_image" "cos" {
    family = "cos-stable"
    project = "cos-cloud"
}

module "template_id" {
    source = "../../random_id"

    byte_length = 4
    keepers = {
        // Depend on all the variables used by google_compute_instance_template.template
        // so that we get a new random_id each time the variables to this module are
        // changed by the user. This is necessary because template names must be unique.
        //
        // Compound values below use hashes to reduce redundancy in diffs (actual changes
        // to these values will be shown elsewhere in the diff).
        name = var.name
        tags = base64sha512(join(" ", var.tags))
        container_image = base64sha512(jsonencode(var.container_image))
        machine_type = var.machine_type
        host_to_container_ports = base64sha512(jsonencode(var.host_to_container_ports))
        envoy_config = base64sha512(jsonencode(var.envoy_config))
        args = base64sha512(jsonencode(var.args))
        env = base64sha512(jsonencode(var.env))
        network = var.network
        subnetwork = var.subnetwork
        public_ip = var.public_ip
        preemptible = var.preemptible
        service_account = var.service_account
        templatefile = filebase64sha512("${path.module}/gce_cloud-init.tmpl.yaml")
    }
}

resource "google_compute_instance_template" "template" {
    lifecycle {
        create_before_destroy = true
    }

    name = "${var.name}-${module.template_id.id}"
    // name_prefix produces really long template names
    tags = length(var.tags) > 0 ? var.tags : null
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
    machine_type = var.machine_type
    disk {
        source_image = data.google_compute_image.cos.self_link
        auto_delete = true
        boot = true
    }
    metadata = {
        user-data = templatefile("${path.module}/gce_cloud-init.tmpl.yaml",
            {
                service_name = var.name
                container_image_name = var.container_image.image_url
                host_to_container_ports = var.host_to_container_ports
                args = var.args != null ? var.args : []
                env = var.env != null ? var.env : {}
                envoy_config = var.envoy_config
            }
        )
        google-logging-enabled = "true"
        google-monitoring-enabled = "true"
        cos-metrics-enabled = "true"
        enable-oslogin = "true"
    }

    network_interface {
        network = (var.network == null && var.subnetwork == null) ? "default" : var.network
        subnetwork = var.subnetwork

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

    dynamic "service_account" {
        for_each = var.service_account != null ? [1] : []
        content {
            email = var.service_account
            scopes = [
                // Restrict via IAM on service account:
                // https://cloud.google.com/compute/docs/access/service-accounts#service_account_permissions
                "https://www.googleapis.com/auth/cloud-platform",
            ]
        }
    }

    shielded_instance_config {
      enable_secure_boot = true
      enable_vtpm = true
      enable_integrity_monitoring = true
    }
}
