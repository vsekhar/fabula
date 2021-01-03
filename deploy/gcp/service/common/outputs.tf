output "regional_health_check_id" {
    value = google_compute_region_health_check.hc.id
}

output "instance_group" {
    value = google_compute_region_instance_group_manager.rigm.instance_group
}
