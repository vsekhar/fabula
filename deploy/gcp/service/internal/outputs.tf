output "self_link" {
    value = try("http://${google_compute_forwarding_rule.forwarding_rule[0].ip_address}", null)
}

output "ip" {
    value = try(google_compute_forwarding_rule.forwarding_rule[0].ip_address, null)
}

output "service_name" {
    value = try(google_compute_forwarding_rule.forwarding_rule[0].service_name, null)
}
