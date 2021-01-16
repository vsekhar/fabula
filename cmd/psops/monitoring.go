package main

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	latencyMs = stats.Int64("publish_latency", "pubsub publish latency in milliseconds", "ms")
)

func billableSize(m *pubsub.Message) int {
	// Source: https://cloud.google.com/pubsub/pricing#message_delivery_pricing
	r := 0
	r += len(m.Data)
	for k, v := range m.Attributes {
		r += len(k) + len(v)
	}
	r += 20 // timestamp
	r += 16 // message_id
	r += len(m.OrderingKey)
	return r
}

func initMonitoring(_ context.Context) {
	v := &view.View{
		Name:        "publish_latency_distribution",
		Measure:     latencyMs,
		Description: "The distribution of pubsub publish latencies",

		Aggregation: view.Distribution(0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100),
	}
	if err := view.Register(v); err != nil {
		log.Fatal(err)
	}
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer exporter.Flush()
	if err := exporter.StartMetricsExporter(); err != nil {
		log.Fatal(err)
	}
	defer exporter.StopMetricsExporter()
	view.RegisterExporter(exporter)
}
