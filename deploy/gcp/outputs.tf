// TODO: server/endpoint URLs

output "self_link" {
    value = "http://${google_compute_instance_from_template.fabula-instance.network_interface.0.access_config.0.nat_ip}"
}
