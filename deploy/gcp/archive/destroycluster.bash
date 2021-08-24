#!/bin/env bash

set -ex

# Just destroy the cluster, then state rm the dependent resources,
# otherwise Terraform will try to destroy the k8s namespace and it takes
# forever.

terraform destroy -target=google_container_cluster.fabula
terraform state rm module.fabula_kubernetes.kubernetes_namespace.fabula
# TODO: add more entries based on what Terraform reports is missing after
# running this script and then trying to recreate the cluster.
