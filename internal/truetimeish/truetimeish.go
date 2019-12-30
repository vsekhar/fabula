// Package truetimeish emulates a TrueTime service.
package truetimeish

import (
	"time"
)

const epsilon = 10 * time.Millisecond

func nowish() (earliest, latest time.Time) {
	t := time.Now()
	return t.Add(-epsilon), t.Add(epsilon)
}

func sleepUntilPast(t time.Time) {
	e, _ := nowish()
	for e.Sub(t) > 0 {
		time.Sleep(e.Sub(t))
		e, _ = nowish()
	}
}

// Request is a deferred request for a causal timestamp.
type Request struct {
	t    time.Time
	past bool
}

// Timestamp blocks until the deferred timestamp is in the past, and then
// returns its value.
func (r *Request) Timestamp() time.Time {
	if !r.past {
		sleepUntilPast(r.t)
		r.past = true
	}
	return r.t
}

// Get returns a deferred request for a causal timestamp.
//
// Deferring the request permits clients to perform concurrent work during the
// commit-wait period. Clients call Timestamp on the TimestampRequest when the
// actual timestamp value is required.
//
// For performance, clients should call Request as early in the interval within
// which the timestamp is required (e.g. as soon as all locks are acquired), and
// clients should call Timestamp on the returned request as late as possible before
// the timestamp value is required. Doing so reduces the sleep required to ensure
// timestamp causality.
func Get() Request {
	return Request{t: time.Now(), past: false}
}
