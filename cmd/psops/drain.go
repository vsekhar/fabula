package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
)

func drain(ctx context.Context) {
	startTime := time.Now()
	sub := getSubscriptionOrCreateOrDie(ctx, "all")
	var count int64
	var bytes uint64
	rctx, cancel := context.WithCancel(ctx)
	cancelOnce := sync.Once{}
	bar := progressbar.NewOptions(*n, pbopts()...)
	bar.Describe("Draining entries...")
	bar.RenderBlank()
	err := sub.Receive(rctx, func(ctx context.Context, msg *pubsub.Message) {
		c := atomic.AddInt64(&count, 1)
		if *n > 0 && c > int64(*n) {
			atomic.AddInt64(&count, -1)
			cancelOnce.Do(func() {
				cancel()
				msg.Nack()
				return
			})
		}
		atomic.AddUint64(&bytes, uint64(billableSize(msg)))
		bar.Add(1)
		msg.Ack()
		if *verbose {
			log.Printf("Drained %s... (OrderingKey: %s, PublishTime: %s)", enc(msg.Data[:8]), msg.OrderingKey, msg.PublishTime)
		}
	})
	bar.Finish()
	elapsed := time.Now().Sub(startTime)
	rate := float64(count) / elapsed.Seconds()
	dataRate := bytes / uint64(elapsed.Seconds())
	log.Printf("Drained %s messages, %s in %s (%s entries/s, %s/s)\n",
		humanize.Comma(count), humanize.Bytes(bytes),
		elapsed, humanize.FormatFloat("", rate),
		humanize.Bytes(dataRate))
	if err != nil {
		log.Print(err)
	}
}
