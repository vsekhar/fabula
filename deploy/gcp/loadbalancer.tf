resource "google_compute_global_address" "default" {
  name         = "${var.project_id}-public-address"
  ip_version   = "IPV4"
  address_type = "EXTERNAL"
}
