package main

// Fabula uses regions to improve latency to users. Each region has its own
// Merkle weave.

// regionMap maps GCP regions to Fabula regions.
var regionMap map[string]string = map[string]string{
	"us-east1":                "NA",
	"us-east4":                "NA",
	"northamerica-northeast1": "NA",
	"southamerica-east1":      "NA",
	"us-central1":             "NA",
	"us-west1":                "NA",
	"us-west2":                "NA",
	"us-west3":                "NA",
	"us-west4":                "NA",

	"europe-west1": "EU",
	"europe-west2": "EU",
	"europe-west3": "EU",
	"europe-west4": "EU",
	"europe-west6": "EU",
}
