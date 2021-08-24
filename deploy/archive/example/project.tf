// Create a new project and deploy fabula into it.

provider "google" {
    version = "~> 3.50.0"
}

variable "fabula_region" {
    type = string
}

variable "fabula_project" {
    type = string
}

module "regions" {
    source = "../regions"
    fabula_region = var.fabula_region
}

variable "billing_account" {
    type = string
}

resource "random_id" "id" {
    byte_length = 4
    prefix = "fabula-dev-"
}

provider "google" {
    version = "~> 3.50.0"
    project = var.fabula_project
    region = module.regions.provider_region
    alias = "for_fabula"
}

module "fabula" {
    source = "../"
    fabula_region = var.fabula_region
    project_id = var.fabula_project
    providers = {
        google = google.for_fabula
    }
}
