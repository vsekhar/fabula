resource "google_project_iam_member" "logging-fabula" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.fabula.email}"
}

resource "google_project_iam_member" "monitoring-fabula" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.fabula.email}"
}

resource "google_project_iam_member" "logging-storer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.storer.email}"
}

resource "google_project_iam_member" "monitoring-storer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.storer.email}"
}
