module "regions" {
    source = "../../regions"
    fabula_region = var.fabula_region
}

provider "google" {
    version = "~> 3.51.0"
    project = var.project_id
}

locals {
    lifecycle_rules = [
        {
            age = 30
            storage_class = "NEARLINE"
        },
        {
            age = 90
            storage_class = "COLDLINE"
        },
        {
            age = 365
            storage_class = "ARCHIVE"
        }
    ]
}

resource "google_storage_bucket" "pack_storage" {
    name = "${var.project_id}-pack_storage" // global namespace
    location = module.regions.bucket_region

    // For terraform, separate from bucket lifecycle_rule below.
    lifecycle {
        prevent_destroy = true
    }

    dynamic "lifecycle_rule" {
        for_each = local.lifecycle_rules
        content {
            condition {
                age = lifecycle_rule.value["age"]
            }
            action {
                type = "SetStorageClass"
                storage_class = lifecycle_rule.value["storage_class"]
            }
        }
    }
}

resource "google_storage_bucket" "public_storage" {
    name = "${var.project_id}-public_storage" // global namespace
    location = module.regions.bucket_region
    requester_pays = true

    // For terraform, separate from bucket lifecycle_rule below.
    lifecycle {
        prevent_destroy = true
    }

    dynamic "lifecycle_rule" {
        for_each = local.lifecycle_rules
        content {
            condition {
                age = lifecycle_rule.value["age"]
            }
            action {
                type = "SetStorageClass"
                storage_class = lifecycle_rule.value["storage_class"]
            }
        }
    }
}
