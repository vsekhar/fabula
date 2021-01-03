resource "google_compute_region_instance_group_manager" "rigm" {
    name = "service-${var.name}-rigm"
    base_instance_name = "service-${var.name}-instances"
    region = var.region
    auto_healing_policies {
        health_check = var.global_health_check
        initial_delay_sec = 180
    }
    dynamic "version" {
        for_each = var.versions
        content {
            name = version.value["name"]
            instance_template = version.value["instance_template"]
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

resource "google_compute_region_backend_service" "be" {
    name = "service-${var.name}-be"
    region = var.region
    health_checks = var.region_health_checks

    // TODO: internal and exteranl options
    load_balancing_scheme = "EXTERNAL"

    // TODO: separate rigms for each version?
    backend {
      group = google_compute_region_instance_group_manager.rigm.instance_group
    }
}

resource "google_compute_region_autoscaler" "autoscaler" {
    name = "service-${var.name}-autoscaler"
    provider = google-beta // for filter and single_instance_assignment
    region = var.region
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

resource "google_compute_forwarding_rule" "forwarding_rule" {
    name = "service-${var.name}-forwarding-rule"
    region = var.region
    backend_service = google_compute_region_backend_service.be.id
    port_range = "80"
}
