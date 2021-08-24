output "self_link" {
    value = "http://${google_compute_forwarding_rule.forwarding_rule.ip_address}"
}

output "ip" {
    value = google_compute_forwarding_rule.forwarding_rule.ip_address
}

output "service_name" {
    value = google_compute_forwarding_rule.forwarding_rule.service_name
}
