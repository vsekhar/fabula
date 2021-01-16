resource "google_service_account" "storer" {
    account_id = "storer-service-account"
    display_name = "Fabula storer service account"
}

data "google_container_registry_image" "storer_latest" {
    project = "fabula-resources"
    name = "storer"
    tag = "latest"
    // digest = chomp(file("../containers/storer-digest.txt"))
}

module "storer_service" {
    source = "./service/pubsub"
    name = "storer"
    group = module.fabula_group
    min_replicas = 1
    max_replicas = 2
    http_health_check_path = "/healthz/readiness"
    http_health_check_port = 8081
    service_to_container_ports = {
        "8081" = "8081"
    }

    versions = {
        "stable" = {
            container_image = data.google_container_registry_image.storer_latest
            machine_type = "e2-small"
            preemptible = true
            args = [
                "-pubsubSubscription ${google_pubsub_subscription.storer.name}",
                "-bucket ${data.terraform_remote_state.bucket.outputs.pack_storage_name}",
                "-folder batches",
                "-prefix fabula-batch-",
                "-healthCheckPort 8081",
            ]
            service_account = google_service_account.storer.email
        }
    }

    pubsub_autoscale = {
        subscription = google_pubsub_subscription.storer.name
        single_instance_assignment = 15
    }
}
