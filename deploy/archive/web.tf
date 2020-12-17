resource "google_service_account" "webserver" {
    account_id = "webserver"
    display_name = "Webserver"
}

module "gce-container" {
    source = "terraform-google-modules/container-vm/google"
    version = "~> 2.0"
    container = {
        // image = local.web.image
        image = "gcr.io/google-samples/hello-app:1.0"
    }
}

module "gce-container-canary" {
    source = "terraform-google-modules/container-vm/google"
    version = "~> 2.0"
    container = {
        // image = local.web.image
        image = "gcr.io/google-samples/hello-app:2.0"
    }
}

resource "google_compute_network" "webnetwork" {
    name = "webnetwork"
    auto_create_subnetworks = false
}
resource "google_compute_subnetwork" "websubnetwork" {
    name = "websubnetwork"
    ip_cidr_range = "10.125.0.0/20"
    network = google_compute_network.webnetwork.self_link
    region = module.regions.gce_region
    private_ip_google_access = true
}

// Router and Cloud NAT are required for installing packages from repos
// https://github.com/terraform-google-modules/terraform-google-container-vm/blob/master/examples/managed_instance_group/main.tf#L50
// TODO: add if needed

module "web_instance_templates" {
    source = "terraform-google-modules/vm/google//modules/preemptible_and_regular_instance_templates"

    network = google_compute_network.webnetwork.self_link
    subnetwork = google_compute_subnetwork.websubnetwork.self_link
    service_account = {
        email = google_service_account.webserver.email
        scopes = [] // legacy but required
    }
    name_prefix = "fabula-web"
    source_image_family = "cos-stable"
    source_image_project = "cos-cloud"
    source_image = reverse(split("/", module.gce-container.source_image))[0]
    metadata = map("gce-container-declaration", module.gce-container.metadata_value)
    tags = [
        "container-vm-webserver-mig"
    ]
    labels = {
        "container-vm" = module.gce-container.vm_container_label
    }
}

resource "google_compute_health_check" "web_health_check" {
    name = "web-health-check"
    timeout_sec = 1
    check_interval_sec = 1
    http_health_check {
        request_path = "/healthz"
        port = 8080
    }
}

resource "google_compute_target_pool" "web_pool" {
    name = "web-pool"
    health_checks = [
        google_compute_health_check.web_health_check.name
    ]
}

// TODO: regular (non-preemptible) mig and autoscaler

resource "google_compute_region_instance_group_manager" "web_mig_preemptible" {
    name = "web-mig"
    base_instance_name = "web"
    region = module.regions.gce_region

    version {
        instance_template = module.web_instance_templates.preemptible_self_link
    }

    target_pools = [
        google_compute_target_pool.web_pool.id
    ]

    named_port {
        name = "www"
        port = 8080
    }

    named_port {
        name = "control"
        port = 7946
    }
}

resource "google_compute_region_autoscaler" "autoscaler" {
    name = "autoscaler"
    region = module.regions.gce_region
    target = google_compute_region_instance_group_manager.web_mig_preemptible.id
    autoscaling_policy {
        max_replicas = 5
        min_replicas = 1
        cooldown_period = 60
        // cpu_utilization = 0.6
        metric {
            name = "www.googleapis.com/compute/instance/network/received_bytes_count"
            type = "DELTA_PER_MINUTE"
        }
    }
}

module "http_lb" {
    source = "GoogleCloudPlatform/lb-http/google"
    version = "~> 4.4"

    name = "lb"
    project = var.project_id // not sure why this doesn't default to the provider project
    firewall_networks = [
        google_compute_network.webnetwork.self_link
    ]
    target_tags = [
        "container-vm-webserver-mig"
    ]
    backends = {
        default = {
            description = null
            protocol = "HTTP"
            port = 80
            port_name = "http"
            enable_cdn = true // true doesn't seem to work
            quic = true
            affinity_cookie_ttl_sec = null
            connection_draining_timeout_sec = null
            custom_request_headers = null
            health_check = {
                check_interval_sec = null
                timeout_sec = 10
                healthy_threshold = null
                unhealthy_threshold = null
                request_path = "/"
                port = 80
                host = null
                logging = null
            }
            log_config = {
                enable = true
                sample_rate = 1.0 // TODO: lower
            }
            security_policy = null
            session_affinity = null
            timeout_sec = 10

            groups = [
                {
                    group = google_compute_region_instance_group_manager.web_mig_preemptible.instance_group
                    balancing_mode = null
                    capacity_scaler = null
                    description = null
                    max_connections = null
                    max_connections_per_instance = null
                    max_connections_per_endpoint = null
                    max_rate = null
                    max_rate_per_instance = null
                    max_rate_per_endpoint = null
                    max_utilization = null
                },
            ]
            iap_config = {
                enable = false
                oauth2_client_id = ""
                oauth2_client_secret = ""
            }
        }
    }
}
