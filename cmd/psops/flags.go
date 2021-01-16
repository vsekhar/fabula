package main

import (
	"flag"
	"time"
)

var (
	n       = flag.Int("n", 10, "stop after n entries (0 means go until interrupted)")
	p       = flag.Int("p", 100, "number of concurrent tasks with which to publish")
	t       = flag.String("t", "fabula", "PubSub topic to use")
	region  = flag.String("region", "us-central1", "region")
	verbose = flag.Bool("verbose", false, "verbose output")
	project = flag.String("project", "fabula-rasa-2", "GCP project")
	timeout = flag.Duration("timeout", time.Duration(0), "timeout (0 means go until interrupted)")
	bucket  = flag.String("bucket", "fabula-nam", "GCS bucket to pack into")
)
