// Buckets are defined and managed separately to prevent accidental
// deletion. Before applying the gcp config, run
//
//   $ cd bucket && terraform apply -var-file=../terraform.tfvars
data "terraform_remote_state" "bucket" {
    backend = "local"
    config = {
        path = "bucket/terraform.tfstate"
    }
}

resource "google_storage_bucket_iam_member" "fabula_create" {
    bucket = data.terraform_remote_state.bucket.outputs.pack_storage_name
    role = "roles/storage.objectCreator"
    member = "serviceAccount:${google_service_account.storer.email}"
}
