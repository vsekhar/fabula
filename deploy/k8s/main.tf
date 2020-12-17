resource "kubernetes_namespace" "fabula" {
  metadata {
    generate_name = "fabula-"
  }
}

resource "kubernetes_service_account" "fabula" {
    metadata {
        generate_name = "fabula-k8s-service-account-"
        namespace = kubernetes_namespace.fabula.metadata[0].name
        annotations = {
            "iam.gke.io/gcp-service-account" = var.service_account_email
        }
    }
}

resource "kubernetes_role" "fabula_k8s_role" {
    metadata {
        generate_name = "fabula-k8s-role-"
        namespace = kubernetes_namespace.fabula.metadata[0].name
    }

    rule {
        api_groups = [""]
        resources = ["pods"]
        verbs = ["get", "list", "watch"]
    }
    rule {
        api_groups = ["apps"]
        resources = ["deployments"]
        verbs = ["get", "list"]
    }
}

resource "kubernetes_role_binding" "fabula_k8s_role_binding" {
    metadata {
        // https://github.com/hashicorp/terraform-provider-kubernetes/issues/588
        // generate_name = "fabula-k8s-role-binding-"
        name = "fabula-k8s-role-binding"
        namespace = kubernetes_namespace.fabula.metadata[0].name
    }
    role_ref {
        api_group = "rbac.authorization.k8s.io"
        kind = "Role"
        name = kubernetes_role.fabula_k8s_role.metadata[0].name
    }
    subject {
        kind = "ServiceAccount"
        name = kubernetes_service_account.fabula.metadata[0].name
        namespace = kubernetes_namespace.fabula.metadata[0].name
    }
}

// TODO: fabula seed deployment

resource "kubernetes_deployment" "fabula" {
    metadata {
        name = "fabula-deployment"
        namespace = kubernetes_namespace.fabula.metadata[0].name
        labels = {
            app = "fabula"
        }
    }

    lifecycle {
        ignore_changes = [
            spec[0].replicas, # managed by HPA below
        ]
    }

    spec {
        selector {
            match_labels = {
                app = "fabula"
            }
        }

        template {
            metadata {
                namespace = kubernetes_namespace.fabula.metadata[0].name
                labels = {
                    app = "fabula"
                }
            }

            spec {
                service_account_name = kubernetes_service_account.fabula.metadata[0].name
                automount_service_account_token = true

                container {
                    name = "fabula"
                    image = var.fabula_image
                    args = [
                        "-port=8080",
                        "-notarizerpcport=18193",
                        "-packrpcport=28193",
                        "-controlport=7946", // serf default
                        "-bucket=${var.storage_bucket_name}",
                        "-usereventperiod=5s",
                        "-verbose",
                    ]

                    // Provide pod information for P2P peering
                    env {
                        name = "K8S_NAMESPACE"
                        value_from {
                            field_ref {
                                field_path = "metadata.namespace"
                            }
                        }
                    }
                    env {
                        name = "K8S_POD_NAME"
                        value_from {
                            field_ref {
                                field_path = "metadata.name"
                            }
                        }
                    }

                    port {
                        container_port = 18193
                    }

                    liveness_probe {
                        http_get {
                            path = "/_liveness"
                            port = 8080
                        }
                        initial_delay_seconds = 3
                        period_seconds = 3
                    }
                }
                container {
                    name = "esp"
                    image = "gcr.io/endpoints-release/endpoints-runtime:2"
                    image_pull_policy = "Always"

                    port {
                        container_port = 8081
                    }
                    args = [
                        "--listener_port", "8081",
                        "--backend", "grpc://127.0.0.1:18193",
                        "--service", var.endpoints_service_name,
                        "--rollout_strategy", "managed",
                    ]
                }
            }
        }
    }
}

resource "kubernetes_horizontal_pod_autoscaler" "fabula" {
    metadata {
        name = "fabula"
        namespace = kubernetes_namespace.fabula.metadata[0].name
    }

    spec {
        min_replicas = 2 # for testing discovery
        max_replicas = 10

        scale_target_ref {
            api_version = "apps/v1"
            kind = "Deployment"
            name = kubernetes_deployment.fabula.metadata[0].name
        }

        # Terraform is order-sensitive when diffing against API response, so
        # rejig order of metrics here if you're getting perpetual diffs.
        metric {
            type = "Resource"
            resource {
                name = "memory"
                target {
                    type = "Utilization"
                    average_utilization = "60"
                }
            }
        }
        metric {
            type = "Resource"
            resource {
                name = "cpu"
                target {
                    type = "Utilization"
                    average_utilization = "80"
                }
            }
        }
    }
}

resource "kubernetes_service" "fabula" {
    metadata {
        name = "fabula"
        namespace = kubernetes_namespace.fabula.metadata[0].name
        labels = {
            app = "fabula"
        }
    }

    spec {
        selector = {
            app = kubernetes_deployment.fabula.spec.0.template.0.metadata[0].labels.app
        }
        port {
            port = 80
            target_port = 8081 # ESP
        }
        type = "LoadBalancer"
    }
}
