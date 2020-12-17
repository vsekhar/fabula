// These locals map a var.fabula_region to regions for specific resources.
// see locations.md for info

locals {
    // keep in sync with api/service.proto
    fabula_regions = [
        "US",
        "EU",
        "JP",
        "SIN",
        "AUS",
        "IN",
    ]

    // var.fabula_region --> bucket_region
    // NB: we use multi-region buckets here, see storage/storage.md
    bucket_regions = {
        US = "US"
        EU = "EU"
        JP = "APAC"
        SIN = "APAC"
        AUS = "APAC"
        IN = "APAC"
    }

    // var.fabula_region --> gce_region
    gce_regions = {
        US = "us-central1"            // Iowa
        EU = "europe-west1"           // Belgium
        JP = "asia-northeast2"        // Osaka
        SIN = "asia-southeast1"       // Singapore
        AUS = "australia-southeast1"  // Sydney
        IN = "asia-south1"            // Mumbai
    }
}
