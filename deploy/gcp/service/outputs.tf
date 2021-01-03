output "self_link" {
    value = "http://${google_compute_forwarding_rule.forwarding_rule.ip_address}"
}
