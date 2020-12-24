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
    port_range = "80"
}
