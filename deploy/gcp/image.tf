locals {
  image_project = "fabula-resources"
  image_name = "fabula-image"
  image_digest = chomp(file("../server/fabula-image-latest.txt"))
}

data "google_container_registry_image" "fabula" {
  project = local.image_project
  name = local.image_name
  digest = local.image_digest
}
