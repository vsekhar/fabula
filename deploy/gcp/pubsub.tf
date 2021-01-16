resource "google_pubsub_topic" "storer" {
  name = "storer"
}

resource "google_pubsub_subscription" "storer" {
  name  = "storer"
  topic = google_pubsub_topic.storer.name
  ack_deadline_seconds = 20
  expiration_policy {
    ttl = "" // never expire
  }
}

resource "google_pubsub_topic_iam_member" "packer" {
  topic = google_pubsub_topic.storer.name
  role = "roles/pubsub.publisher"
  member = "serviceAccount:${google_service_account.packer.email}"
}

resource "google_pubsub_subscription_iam_member" "storer" {
  subscription = google_pubsub_subscription.storer.name
  role = "roles/pubsub.subscriber"
  member = "serviceAccount:${google_service_account.storer.email}"
}
