// Package faketime provides HTTP handlers to serve fake TrueTime intervals with
// some random jitter added around the system clock.
package faketime

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/vsekhar/COMMIT/pkg/commit"
)

type interval struct {
	earliest, latest int64
}

// epislon returns a random positive duration drawn from a rectified normal
// distribution with mean 0 and standard deviation jitterMs.
func epsilon(jitterMs float64) time.Duration {
	jitterNano := jitterMs * 1000
	return time.Duration(math.Abs(rand.NormFloat64() * jitterNano))
}

// ttNow returns the earliest and latest possible time, as well as the epsilon
// used to calculate the range.
func ttNow(jitterMs float64) (earliest, latest time.Time, eps time.Duration) {
	t := time.Now()
	e := epsilon(jitterMs)
	return t.Add(-e), t.Add(e), e
}

// wait obtains the current time and sleeps until future timestamps are guaranteed
// to be greater than the current time.
func wait(jitterMs float64) (t time.Time, e time.Duration) {
	t, e = time.Now(), epsilon(jitterMs)
	time.Sleep(e)
	return
}

// NowHandler returns an http.HandlerFunc that serves a fake "now" interval.
func NowHandler(jitterMs float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		earliest, latest, epsilon := ttNow(jitterMs)
		w.Header().Add(commit.EarliestHeader, fmt.Sprintf("%d", earliest.UnixNano()))
		w.Header().Add(commit.LatestHeader, fmt.Sprintf("%d", latest.UnixNano()))
		w.Header().Add(commit.EpsilonDebugHeader, fmt.Sprintf("%d", epsilon.Nanoseconds()))
		w.Header().Add(commit.FakeDebugHeader, "")
		// TODO: signature
	}
}

// CommitWaitHandler returns an http.HandlerFunc that serves a fake post-commit wait
// timestamp.
func CommitWaitHandler(jitterMs float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		t, e := wait(jitterMs)
		w.Header().Add(commit.TimestampHeader, fmt.Sprintf("%d", t.UnixNano()))
		w.Header().Add(commit.EpsilonDebugHeader, fmt.Sprintf("%d", e.Nanoseconds()))
		w.Header().Add(commit.FakeDebugHeader, "")
		// TODO: signature
	}
}
