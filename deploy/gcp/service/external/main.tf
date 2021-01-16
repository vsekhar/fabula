module "service_common" {
    source = "../common"
    name = var.name
    group = var.group
    versions = var.versions
    http_health_check_path = var.http_health_check_path
    http_health_check_port = var.http_health_check_port
    min_replicas = var.min_replicas
    max_replicas = var.max_replicas
    service_to_container_ports = var.service_to_container_ports
}

resource "google_compute_region_backend_service" "be" {
    for_each = length(var.service_to_container_ports) > 0 ? {0:""} : {}

    name = "svc-${var.group.name}-${var.name}"
    health_checks = [module.service_common.regional_health_check_id]
    load_balancing_scheme = "EXTERNAL"
    backend {
      group = module.service_common.instance_group
    }
}

resource "google_compute_forwarding_rule" "forwarding_rule" {
    for_each = length(var.service_to_container_ports) > 0 ? {0:""} : {}

    name = "svc-${var.group.name}-${var.name}"
    backend_service = google_compute_region_backend_service.be[0].id
    port_range = "1-65535" // prevent perpetual diffs which force replacement
}

resource "google_compute_firewall" "allow-external" {
    for_each = length(var.service_to_container_ports) > 0 ? {0:""} : {}

    name = "svc-${var.group.name}-${var.name}-allow-external"
    network = var.group.network
    allow {
        protocol = "tcp"
        ports = [for k, v in var.service_to_container_ports : k]
    }
}
