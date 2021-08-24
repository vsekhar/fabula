// Package batcher supports batching of items. Batching amortizes an action with
// with fixed costs over multiple items. For example, if an API provides an RPC
// that accepts a list of items as input, but clients would prefer adding items
// one at a time, then a Bundler can accept individual items from the client and
// bundle many of them into a single RPC.
//
// The semantics of package batcher are similar to those of package
//
//   google.golang.org/api/support/bundler
//
// except that package batcher aims to eliminate any additional latency in the
// processing of items. That is, batcher will ensure there is a handler running
// to as soon as the first item is available to be handled while also ensuring
// that concurrently available items get batched. In contrast, package bundler
// either waits for items to arrive (adding to item latency) or produces bundles
// of size 1.
package batcher

import (
	"context"
	"reflect"
	"time"
)

const handlerManagementPeriod = 5 * time.Second
const handlerTimeout = 5 * time.Second

// A Batcher collects items added to it into a batch, then calls a user-provided
// function to handle the batch.
type Batcher struct {
	ch            chan interface{}
	handler       func(interface{})
	itemSliceZero reflect.Value // nil (zero value) for slice of items

	// slots for handler invocations
	tokenBucket  chan struct{} // buffered channel loaded with handlerLimit
	maxBatchSize int
}

// New creates a new Batcher.
//
// itemExample is a value of the type that will be batched. For example, if you
// want to create batches of *Entry, you could pass &Entry{} for itemExample.
// Batches will be at most maxBatchSize.
//
// handler is a function that will be called on each bundle. If itemExample is
// of type T, the argument to handler is of type []T. handler may be called
// multiple times concurrently up to maxConcurrentHandlers.
func New(itemExample interface{}, handler func(interface{}), maxBatchSize int, maxConcurrentHandlers int) *Batcher {
	b := &Batcher{
		ch:            make(chan interface{}),
		handler:       handler,
		itemSliceZero: reflect.Zero(reflect.SliceOf(reflect.TypeOf(itemExample))),
		tokenBucket:   make(chan struct{}, maxConcurrentHandlers),
		maxBatchSize:  maxBatchSize,
	}
	for i := 0; i < maxConcurrentHandlers; i++ {
		b.tokenBucket <- struct{}{}
	}
	go b.handleBatch()
	return b
}

// handleBatch breates a batch and handles it. It also starts the next instance
// if handleBatch.
func (b *Batcher) handleBatch() {
	<-b.tokenBucket
	defer func() {
		b.tokenBucket <- struct{}{}
	}()

	batch := b.itemSliceZero
loop:
	for {
		// collect until we have something AND (there's nothing waiting OR
		// our batch is full)
		if batch.Len() < b.maxBatchSize {
			select {
			case v := <-b.ch:
				batch = reflect.Append(batch, reflect.ValueOf(v))
			default:
				if batch.Len() > 0 {
					break loop
				}
				// block until we get something then try again
				v := <-b.ch
				batch = reflect.Append(batch, reflect.ValueOf(v))
			}
		}
	}

	go b.handleBatch()
	b.handler(batch.Interface())
}

// Add adds item to be batched. The type of item must be assignable to the
// itemExample parameter of the NewBundler method, otherwise there will be a
// panic.
func (b *Batcher) Add(ctx context.Context, item interface{}) error {
	// block until timeout
	select {
	case b.ch <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
