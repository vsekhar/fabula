module "service_common" {
    source = "../common"
    name = var.name
    group = var.group
    versions = var.versions
    http_health_check_path = var.http_health_check_path
    http_health_check_port = var.http_health_check_port
    min_replicas = var.min_replicas
    max_replicas = var.max_replicas
    pubsub_autoscale = var.pubsub_autoscale
    service_to_container_ports = var.service_to_container_ports
}

resource "google_compute_region_backend_service" "be" {
    name = "svc-${var.group.name}-${var.name}-be"
    health_checks = [module.service_common.regional_health_check_id]
    load_balancing_scheme = "EXTERNAL"

    // TODO: separate rigms for each version?
    backend {
      group = module.service_common.instance_group
    }
}

resource "google_compute_forwarding_rule" "forwarding_rule" {
    name = "svc-${var.group.name}-${var.name}-forwarding-rule"
    backend_service = google_compute_region_backend_service.be.id
    port_range = "1-65535" // prevent perpetual diffs which force replacement
}

resource "google_compute_firewall" "allow-external" {
    for_each = length(var.service_to_container_ports) > 0 ? {0:""} : {} // dynamic resources need a key

    name = "svc-${var.group.name}-${var.name}-allow-external"
    network = var.group.network
    allow {
        protocol = "tcp"
        ports = [for k, v in var.service_to_container_ports : k]
    }
}
