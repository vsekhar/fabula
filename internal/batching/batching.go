// Package batching provides routines for concurrently grouping large numbers of
// of objects to opreate on them in a batch.
package batching

import (
	"time"
)

// Channel returns a channel that will produce batches of items at most
// batchSize in number and with no item waiting longer than maxWait.
//
// When finished, the caller should close rch. BatchChannel will then close the
// returned channel after the last batch is processed.
func Channel(rch <-chan interface{}, batchSize int, maxWait time.Duration) <-chan []interface{} {
	batches := make(chan []interface{})
	go func() {
		var items []interface{}
		t := time.NewTimer(maxWait)
		if !t.Stop() {
			<-t.C
		}
		for {
			select {
			case <-t.C:
				t.Stop()
				batches <- items
				items = nil
			case n, ok := <-rch:
				if ok {
					// Received new item
					items = append(items, n)
					if len(items) == 1 {
						t.Reset(maxWait)
					}
					if len(items) == batchSize {
						batches <- items
						items = nil
						if !t.Stop() {
							<-t.C
						}
					}
				} else {
					// Channel closed
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
