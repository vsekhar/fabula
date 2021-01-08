locals {
    num_zones = 3
}

module "container_vm_template" {
    for_each = var.versions

    source = "../../container_vm"
    name = "svc-${var.group.name}-${var.name}-${each.key}"
    container_image = each.value["container_image"]
    args = each.value["args"]
    env = each.value["env"]
    host_to_container_ports = var.service_to_container_ports
    envoy_config = each.value["envoy_config"]
    preemptible = each.value["preemptible"]
    machine_type = each.value["machine_type"]
    network = var.group.network
    subnetwork = try(var.group.subnetwork, null)
    service_account = each.value["service_account"]
}

data "google_compute_zones" "available" {}

resource "google_compute_instance_group_manager" "igm" {
    for_each = toset(slice(data.google_compute_zones.available.names, 0, local.num_zones))

    name = "svc-${var.group.name}-${var.name}-${each.value}"
    zone = each.value
    base_instance_name = "svc-${var.group.name}-${var.name}"
    auto_healing_policies {
        health_check = google_compute_health_check.hc.id
        initial_delay_sec = 300 // 5 mins
    }
    dynamic "version" {
        for_each = var.versions
        content {
            name = version.key
            instance_template = module.container_vm_template[version.key].self_link
            dynamic "target_size" {
                for_each = version.value["target_size"] != null ? [1] : []
                content {
                    fixed = version.value["target_size"].fixed
                    percent = version.value["target_size"].percent
                }
            }
        }
    }
}

resource "google_compute_autoscaler" "autoscaler" {
    for_each = toset(slice(data.google_compute_zones.available.names, 0, local.num_zones))

    name = "svc-${var.group.name}-${var.name}-${each.value}"
    provider = google-beta // for filter and single_instance_assignment
    target = google_compute_instance_group_manager.igm[each.key].id
    zone = each.value
    autoscaling_policy {
        mode = "ON"
        max_replicas = var.max_replicas
        min_replicas = 1
        cooldown_period = 120
        cpu_utilization {
            target = 0.6
        }

        // https://cloud.google.com/compute/docs/autoscaler/scaling-stackdriver-monitoring-metrics#example_using_instance_assignment_to_scale_based_on_a_queue
        metric {
            name = "pubsub.googleapis.com/subscription/num_undelivered_messages"
            filter = "resource.type = pubsub_subscription AND resource.label.subscription_id = ${var.pubsub_autoscale.subscription}"

            // There will be local.num_zones operating in parallel, so we treat each instance
            // as having num_zones * single_instance_assignment capacity.
            single_instance_assignment = var.pubsub_autoscale.single_instance_assignment * local.num_zones
        }
    }
}

resource "google_compute_health_check" "hc" {
    name = "svc-${var.group.name}-${var.name}-http"
    http_health_check {
      port = var.http_health_check_port
      request_path = var.http_health_check_path
    }
}

resource "google_compute_firewall" "allow_health_checks" {
    name = "svc-${var.group.name}-${var.name}-allow-health-checks"
    network = var.group.network

    // https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges
    source_ranges = [ "35.191.0.0/16", "130.211.0.0/22" ]
    direction = "INGRESS"
    allow {
        protocol = "tcp"
        ports = [var.http_health_check_port]
    }
}
