locals {
    target_tags = [
        "fabula-container-vm-mig"
    ]
}

data "google_compute_image" "cos" {
    family = "cos-stable"
    project = "cos-cloud"
}

data "template_file" "gce_cos_yaml" {
    template = file("${path.module}/gce_cos.tmpl.yaml")
    vars = {
        container_image_name = data.google_container_registry_image.fabula.image_url
    }
}

resource "google_compute_instance_template" "fabula" {
    name = "fabula-instance-template"
    project = var.project_id
    tags = local.target_tags
    labels = {
        container-vm = data.google_compute_image.cos.name
    }
    machine_type = "f1-micro"
    disk {
        source_image = data.google_compute_image.cos.self_link
        auto_delete = true
        boot = true
    }
    metadata = {
        // TODO: try with cloud-init
        // TODO: start stackdriver logging agent? https://stackoverflow.com/a/58722803/608382
        gce-container-declaration = data.template_file.gce_cos_yaml.rendered
        google-logging-enabled = "true"
        enable-oslogin = "true"
        cos-metrics-enabled = "true"
    }

    network_interface {
        network = google_compute_network.vpc.self_link
        subnetwork = google_compute_subnetwork.subnet.self_link
        access_config {
            network_tier = "PREMIUM"
        }
    }

    scheduling {
        preemptible         = "true"
        on_host_maintenance = "TERMINATE" // required for preemptible
        automatic_restart   = "false"     // required for preemptible
    }

    service_account {
        email = google_service_account.fabula.email
        scopes = []
    }
}

# TODO: make this a mig
# TODO: add autoscaler to mig

resource "google_compute_instance_from_template" "fabula-instance" {
    name = "fabula-instance"
    zone = "us-central1-a"
    source_instance_template = google_compute_instance_template.fabula.id
}

resource "google_compute_target_pool" "fabula" {
    name = "fabula-pool"
    region = module.regions.gce_region
    instances = [
        "us-central1-a/fabula-instance"
    ]
}

resource "google_compute_forwarding_rule" "fabula" {
    name = "fabula-forwarding-rule"
    region = module.regions.gce_region
    target = google_compute_target_pool.fabula.id
    // port_range = "80"
}
