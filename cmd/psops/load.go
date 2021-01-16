package main

import (
	"context"
	"crypto/rand"
	"log"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"github.com/vsekhar/fabula/internal/prefix"
	"go.opencensus.io/stats"
)

func load(ctx context.Context) {
	startTime := time.Now()
	topic := getTopicOrDie(ctx, *t)
	defer topic.Stop()

	bar := progressbar.NewOptions(*n, pbopts()...)
	bar.Describe("Loading entries...")

	var count int64
	var bytes uint64
	doPar(ctx, func(ctx context.Context) {
		var r [64]byte
		rand.Read(r[:])
		msg := &pubsub.Message{
			Data:        r[:],
			OrderingKey: prefix.ToString(r[:], prefix.LengthNibbles),
		}
		pubtime := time.Now()
		res := topic.Publish(ctx, msg)
		_, err := res.Get(ctx) // block until sent
		stats.Record(ctx, latencyMs.M(time.Since(pubtime).Milliseconds()))
		if err != nil {
			if err != context.Canceled && err != context.DeadlineExceeded {
				log.Printf("error publishing %s... (order: %s): %s", enc(msg.Data[:8]), msg.OrderingKey, err)
			}
			topic.ResumePublish(msg.OrderingKey)
			// Normally we'd retry to preserve ordering but
			// we just drop it on the floor here.
			return
		}
		bar.Add(1)
		atomic.AddInt64(&count, 1)
		atomic.AddUint64(&bytes, uint64(billableSize(msg)))
		if *verbose {
			log.Printf("Published %s... (OrderingKey: %s)", enc(msg.Data[:8]), msg.OrderingKey)
		}
	})
	bar.Finish()
	elapsed := time.Now().Sub(startTime)
	rate := float64(count) / elapsed.Seconds()
	dataRate := bytes / uint64(elapsed.Seconds())
	log.Printf("Published %s messages, %s in %s (%s entries/s, %s/s)\n",
		humanize.Comma(count), humanize.Bytes(bytes),
		elapsed, humanize.FormatFloat("", rate),
		humanize.Bytes(dataRate))
}
