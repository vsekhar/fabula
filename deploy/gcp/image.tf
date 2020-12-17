data "google_container_registry_image" "fabula" {
  project = "fabula-resources"
  name = "fabula-image"
  digest = chomp(file("../server/fabula-image-latest.txt"))
}
