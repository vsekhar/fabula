package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
)

func getStorageClientOrDie(ctx context.Context) *storage.Client {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func getBucketOrDie(ctx context.Context, name string) *storage.BucketHandle {
	client := getStorageClientOrDie(ctx)
	return client.Bucket(name)
}
