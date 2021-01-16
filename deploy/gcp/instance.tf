locals {
  image_project = "fabula-resources"
  image_name = "fabula"
  image_digest = chomp(file("../containers/fabula-digest.txt"))
}

locals {
    target_tags = [
        "fabula-container-vm-mig"
    ]
}

data "google_container_registry_image" "fabula" {
  project = local.image_project
  name = local.image_name
  digest = local.image_digest
}

// for testing
data "google_container_registry_image" "hello-app" {
  project = "google-samples"
  name = "hello-app:2.0"
}
