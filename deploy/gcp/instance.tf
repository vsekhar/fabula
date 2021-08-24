
locals {
  image_project = "fabula-resources"
  image_name = "fabula-image"
  image_digest = chomp(file("../server/fabula-image-latest.txt"))
}

locals {
    target_tags = [
        "fabula-container-vm-mig"
    ]
}

data "google_container_registry_image" "fabula" {
  project = local.image_project
  name = local.image_name
  digest = local.image_digest
}

// for testing
data "google_container_registry_image" "hello-app" {
  project = "google-samples"
  name = "hello-app:2.0"
}

module "fabula_instance_template" {
    source = "./container_vm"

    name = "fabula"
    container_image = data.google_container_registry_image.hello-app
    // args = ["-arg1", "-arg2"]
    host_to_container_ports = {
        "80" = "8080"
    }
    network = google_compute_network.vpc.self_link
    subnetwork = google_compute_subnetwork.subnet.self_link
    // public_ip = "0.0.0.0" // ephemeral
    service_account = google_service_account.fabula.email
    preemptible = true
    machine_type = "e2-small"
    tags = local.target_tags
}
