// Package batching provides routines for concurrently grouping large numbers of
// of objects to opreate on them in a batch.
package batching

import (
	"time"
)

const batchBuffer = 100

// Channel returns a channel that will produce batches of items at most
// batchSize in number and with no item waiting longer than maxWait.
//
// When finished, the caller should close rch. BatchChannel will then close the
// returned channel after the last batch is processed.
func Channel(rch <-chan interface{}, batchSize int, maxWait time.Duration) <-chan []interface{} {
	batches := make(chan []interface{}, batchBuffer)
	go func() {
		var items []interface{}
		delayTimer := time.NewTimer(maxWait)
		for {
			select {
			case <-delayTimer.C:
				if len(items) > 0 {
					batches <- items
					items = nil
				}
				delayTimer.Reset(maxWait)
			case n, ok := <-rch:
				if ok {
					items = append(items, n)
					if len(items) == batchSize {
						batches <- items
						items = nil
						delayTimer.Reset(maxWait)
					}
				} else {
					if len(items) > 0 {
						batches <- items
						items = nil
					}
					close(batches)
					return
				}
			}
		}
	}()
	return batches
}
