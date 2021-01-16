resource "google_service_account" "packer" {
    account_id = "packer-service-account"
    display_name = "Fabula packerservice account"
}

// TODO: add roles/compute.viewer (for p2p discovery)
