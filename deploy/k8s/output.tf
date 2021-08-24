output "kubernetes_service_account_id" {
    value = kubernetes_service_account.fabula.id
}

output "kubernetes_load_balancer_ip" {
  value = kubernetes_service.fabula.load_balancer_ingress[0].ip
}
