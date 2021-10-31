package batcher_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/vsekhar/fabula/internal/batcher"
)

func TestBatcher(t *testing.T) {
	// production
	goroutines := 100
	entriesPerGoroutine := 100
	interval := 10 * time.Millisecond
	entries := goroutines * entriesPerGoroutine
	// t = 100 entriesPerGoroutine * 10 ms interval = 1,000 ms
	// 10,000 entries / 1,000 ms = 10,000 entries / s

	// consumption
	// c_interval = ~100ms on average
	// c_rate = 1 / c_interval = ~10 batches per second

	// Expected batch size = (10,000 entries / s) / 0.1 s = 1,000

	handled := 0
	b := batcher.New(int(0), func(i interface{}) {
		batch := i.([]int)
		t.Logf("Batch size %d", len(batch))
		handled += len(batch)
		time.Sleep(time.Duration(87+rand.Intn(25)) * time.Millisecond) // 100ms avg
	}, 10000, 1)
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			for j := 0; j < entriesPerGoroutine; j++ {
				time.Sleep(interval)
				b.Add(ctx, i*j)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	b.Close()
	if handled != entries {
		t.Errorf("expected %d entries handled, got %d", entries, handled)
	}

	// should see small initial batch and batches of ~1000 thereafter
	// t.Error("output")
}
