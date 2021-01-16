package main

import (
	"context"
	"log"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"golang.org/x/oauth2/google"
)

func initTrace(ctx context.Context) {
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if creds.ProjectID == "" {
		log.Fatal("no project ID in default credentials")
	}
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: creds.ProjectID,
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	// TODO: finish this
}
