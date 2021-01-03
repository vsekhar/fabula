output "external_self_link" {
    value = module.external_hello_service.self_link
}

output "external_lb_ip" {
    value = module.external_hello_service.ip
}

output "internal_lb_ip" {
    value = module.internal_hello_service.ip
}
