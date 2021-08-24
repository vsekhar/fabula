module "fabula-service" {
    source = "./service"
    name = "fabula"
    region = module.regions.gce_region
    min_replicas = 1
    network = google_compute_network.vpc
    http_health_check_path = "/"
    http_health_check_port = 80

    versions = [
        {
            name = "prod"
            instance_template = module.fabula_instance_template.self_link
        }
    ]
}
