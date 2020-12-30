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
	var b *batcher.Batcher
	b = batcher.New(int(0), func(i interface{}) {
		batch := i.([]int)
		t.Logf("Batch size %d", len(batch))
		time.Sleep(time.Duration(87+rand.Intn(25)) * time.Millisecond) // 100ms avg
	}, 1000, 100)
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	goroutines := 100
	entriesPerGoroutine := 100

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			for j := 0; j < entriesPerGoroutine; j++ {
				time.Sleep(10 * time.Millisecond) // ~10 entries per handler invocation
				b.Add(ctx, i*j)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	// should see large-ish batch sizes with only a few batches of 1-2 entries.
	// t.Error("output")
}
