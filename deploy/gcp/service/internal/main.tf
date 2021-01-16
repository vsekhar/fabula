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
    load_balancing_scheme = "INTERNAL"
    backend {
      group = module.service_common.instance_group
    }
}

resource "google_compute_forwarding_rule" "forwarding_rule" {
    for_each = length(var.service_to_container_ports) > 0 ? {0:""} : {}

    name = "svc-${var.group.name}-${var.name}"
    network = var.group.network
    subnetwork = var.group.subnetwork
    backend_service = google_compute_region_backend_service.be[0].id
    load_balancing_scheme = "INTERNAL"
    all_ports = true
    // --> lb.svc-<groupname>-<servicename>.il4.<region>.lb.<projectID>.internal
    service_label = "lb"
}

resource "google_compute_firewall" "allow-internal" {
    for_each = length(var.service_to_container_ports) > 0 ? {0:""} : {}

    name = "svc-${var.group.name}-${var.name}-allow-internal"
    network = var.group.network

    // copied from default-allow-internal
    source_ranges = [ "10.128.0.0/9" ]
    priority = 65534
    direction = "INGRESS"
    allow {
        protocol = "icmp"
    }
    allow {
        protocol = "udp"
        ports = [for k, v in var.service_to_container_ports : k]
    }
    allow {
        protocol = "tcp"
        ports = [for k, v in var.service_to_container_ports : k]
    }
}

/*

metadata.items.created-by="projects/523215995107/regions/us-central1/instanceGroupManagers/svc-test-internal"

works:
id="720558720759462875"
name="svc-test-internal-96sv"

*/
