package pubsub

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

const project = "fabula-rasa-2"
const topicName = "fabula"
const subscriptionName = "all"
const region = "us-central1"

func newClientOrDie(ctx context.Context) *pubsub.Client {
	endpoint := "pubsub.googleapis.com:443"
	if region != "" {
		endpoint = region + "-" + endpoint
	}
	client, err := pubsub.NewClient(ctx, project, []option.ClientOption{
		option.WithEndpoint(endpoint),
	}...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func BenchmarkOrderedPublish(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := newClientOrDie(ctx)
	topic := client.Topic(topicName)
	topic.EnableMessageOrdering = true

	// Serialize publication, no batching
	topic.PublishSettings.DelayThreshold = 0
	topic.PublishSettings.CountThreshold = 0
	topic.PublishSettings.ByteThreshold = 0

	var totalLatency time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		res := topic.Publish(ctx, &pubsub.Message{
			Data:        []byte("pubsub_bench_test"),
			OrderingKey: "pubsub_bench_test",
		})
		_, err := res.Get(ctx)
		if err != nil {
			b.Error(err)
		}
		totalLatency += time.Since(start)
	}
	b.StopTimer()
	b.ReportMetric(float64(totalLatency.Milliseconds())/(float64(b.N)), "latency_ms/op")
}
