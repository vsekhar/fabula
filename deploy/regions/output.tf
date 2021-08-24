output "bucket_region" {
    value = local.bucket_regions[var.fabula_region]
}

output "gce_region" {
    value = local.gce_regions[var.fabula_region]
}

output "provider_region" {
    value = local.gce_regions[var.fabula_region]
}
