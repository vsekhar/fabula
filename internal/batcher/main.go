// +build !

package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/vsekhar/fabula/internal/batcher"
)

func main() {
	var b *batcher.Batcher
	handler := func(i interface{}) {
		batch := i.([]int)
		log.Printf("Batch size %d", len(batch))
		time.Sleep(time.Duration(87+rand.Intn(25)) * time.Millisecond) // 100ms avg
	}
	b = batcher.New(int(0), handler, 1000, 50)
	ctx := context.Background()
	wg := new(sync.WaitGroup)
	goroutines := 100
	entriesPerGoroutine := 100
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			for j := 0; j < entriesPerGoroutine; j++ {
				time.Sleep(time.Duration(7+rand.Intn(5)) * time.Millisecond) // 10ms avg
				ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				if err := b.Add(ctx, i*j); err != nil {
					log.Print(err)
				}
				cancel()
			}
			wg.Done()
		}(i)
	}

	wg.Add(goroutines)
	wg.Wait()
}
