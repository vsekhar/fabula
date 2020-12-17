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

resource "google_storage_bucket_iam_binding" "fabula_create" {
    bucket = google_storage_bucket.pack_storage.name
    role = "roles/storage.objectCreator"
    members = [
        "serviceAccount:${google_service_account.fabula.email}",
    ]
}

resource "google_storage_bucket_iam_binding" "fabula_view" {
    bucket = google_storage_bucket.pack_storage.name
    role = "roles/storage.objectViewer"
    members = [
        "serviceAccount:${google_service_account.fabula.email}",
    ]
}

resource "google_storage_bucket" "public_storage" {
    name = "${var.project_id}-public_storage" // global namespace
    location = module.regions.bucket_region
    requester_pays = true

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
