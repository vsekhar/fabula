// TODO: version numbers

locals {
    image_project = "fabula-resources"
    image_path = "gcr.io/${local.image_project}"

    web = {
        image = "${local.image_path}/web:latest"
        port = 8080
    }
}
