// forwarding rule --> be service --> rigm --> instances

resource "google_compute_region_health_check" "http-health-check-be" {
    name = "http-health-check-be"
    region = module.regions.gce_region
    check_interval_sec = 2
    timeout_sec = 1
    http_health_check {
      port = "80"
      request_path = "/"
    }
}

resource "google_compute_health_check" "http-health-check-rigm" {
    name = "http-health-check-rigm"
    check_interval_sec = 2
    timeout_sec = 1
    http_health_check {
      port = "80"
      request_path = "/"
    }
}

resource "google_compute_region_backend_service" "fabula" {
    name = "fabula-be-service"
    region = module.regions.gce_region
    health_checks = [ google_compute_region_health_check.http-health-check-be.id ]
    load_balancing_scheme = "EXTERNAL"

    backend {
      group = google_compute_region_instance_group_manager.fabula.instance_group
    }
}

resource "google_compute_region_instance_group_manager" "fabula" {
    name = "fabula-igm"
    base_instance_name = "fabula"
    region = module.regions.gce_region
    auto_healing_policies {
        health_check = google_compute_health_check.http-health-check-rigm.id
        initial_delay_sec = 180
    }
    version {
        name = "prod"
        instance_template = google_compute_instance_template.fabula.id
    }
    version {
        name = "canary"
        instance_template = google_compute_instance_template.fabula.id
        target_size {
            fixed = 1
        }
    }
}

resource "google_compute_region_autoscaler" "fabula" {
    name = "fabula-autoscaler"
    region = module.regions.gce_region
    target = google_compute_region_instance_group_manager.fabula.id
    autoscaling_policy {
        mode = "ON"
        max_replicas = 9
        min_replicas = 3 // 1 per zone
        cooldown_period = 60
        cpu_utilization {
            target = 0.5
        }
    }
}

resource "google_compute_forwarding_rule" "fabula" {
    name = "fabula-forwarding-rule"
    region = module.regions.gce_region
    backend_service = google_compute_region_backend_service.fabula.id
    port_range = "80"
}
