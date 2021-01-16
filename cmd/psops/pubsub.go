package main

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

var client *pubsub.Client
var clientOnce sync.Once

func getPubsubClientOrDie(ctx context.Context) *pubsub.Client {
	clientOnce.Do(func() {
		endpoint := "pubsub.googleapis.com:443"
		if *region != "" {
			endpoint = *region + "-" + endpoint
		}
		var err error
		client, err = pubsub.NewClient(ctx, *project, []option.ClientOption{
			option.WithEndpoint(endpoint),
		}...)
		if err != nil {
			log.Fatal(err)
		}
	})
	return client
}

func getTopicOrDie(ctx context.Context, name string) *pubsub.Topic {
	client := getPubsubClientOrDie(ctx)
	topic := client.Topic(*t)
	topic.EnableMessageOrdering = true
	return topic
}

func getSubscriptionOrCreateOrDie(ctx context.Context, name string) *pubsub.Subscription {
	client := getPubsubClientOrDie(ctx)
	sub := client.Subscription(name)
	exists, err := sub.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		sub, err = client.CreateSubscription(ctx, name,
			pubsub.SubscriptionConfig{
				Topic:                 getTopicOrDie(ctx, *t),
				EnableMessageOrdering: true,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	return sub
}
