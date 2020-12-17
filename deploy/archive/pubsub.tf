resource "google_pubsub_topic" "pack_topic" {
    name = "pack-topic"
}

resource "google_pubsub_subscription" "pack_subscription" {
    name = "pack-subscription"
    topic = google_pubsub_topic.pack_topic.name
    //
}

// TODO: public subscription permissions of top level topic
