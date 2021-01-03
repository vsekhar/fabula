module "container_vm_template" {
    for_each = var.versions

    source = "../../container_vm"
    name = "svc-${var.group.name}-${var.name}-${each.key}"
    container_image = each.value["container_image"]
    args = each.value["args"]
    env = each.value["env"]
    host_to_container_ports = var.service_to_container_ports
    preemptible = each.value["preemptible"]
    machine_type = each.value["machine_type"]
    network = var.group.network
    subnetwork = try(var.group.subnetwork, null)
    service_account = var.service_account
}

// forwarding rule (int/ext) --> be service (int/ext) --> rigm (common) --> firewall (int/ext) -- > instances (common)
//                                \ region health check   |- autoscaler (common)
//                                  (common)               \ global health check (common)

data "google_client_config" "current" {}

resource "google_compute_region_instance_group_manager" "rigm" {
    name = "svc-${var.group.name}-${var.name}-rigm"
    base_instance_name = "svc-${var.group.name}-${var.name}-inst"
    region = data.google_client_config.current.region // seems to be required for this resource type...
    auto_healing_policies {
        health_check = google_compute_health_check.hc.id
        initial_delay_sec = 180
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

resource "google_compute_region_autoscaler" "autoscaler" {
    name = "svc-${var.group.name}-${var.name}-autoscaler"
    provider = google-beta // for filter and single_instance_assignment
    target = google_compute_region_instance_group_manager.rigm.id
    autoscaling_policy {
        mode = "ON"
        max_replicas = var.max_replicas
        min_replicas = var.min_replicas
        cooldown_period = 120
        cpu_utilization {
            target = 0.6
        }

        // https://cloud.google.com/compute/docs/autoscaler/scaling-stackdriver-monitoring-metrics#example_using_instance_assignment_to_scale_based_on_a_queue
        dynamic "metric" {
            for_each = var.pubsub_autoscale != null ? [1] : []
            content {
                name = "pubsub.googleapis.com/subscription/num_undelivered_messages"
                type = "GAUGE"
                filter = "resource.type = pubsub_subscription AND resource.label.subscription_id = ${var.pubsub_autoscale.subscription}"
                single_instance_assignment = var.pubsub_autoscale.single_instance_assignment
            }
        }
    }
}

resource "google_compute_region_health_check" "hc" {
    name = "svc-${var.group.name}-${var.name}-http-regional"
    http_health_check {
      port = var.http_health_check_port
      request_path = var.http_health_check_path
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
