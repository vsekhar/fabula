// faketime serves fake TrueTime intervals with some random
// jitter added around the system clock.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/vsekhar/COMMIT"
)

var port = flag.Int("port", 8080, "port to listen on (default 8080)")
var jitterMs = flag.Float64("jitter", 10.0, "stddev of timestamp range in milliseconds (default 10.0)")

type interval struct {
	earliest, latest int64
}

// epislon returns a random positive duration drawn from a rectified normal
// distribution with mean 0 and standard deviation *jitterMs.
func epsilon() time.Duration {
	jitterNano := *jitterMs * 1000
	return time.Duration(math.Abs(rand.NormFloat64() * jitterNano))
}

// ttNow returns the earliest and latest possible time, as well as the epsilon
// used to calculate the range.
func ttNow() (earliest, latest time.Time, eps time.Duration) {
	t := time.Now()
	e := epsilon()
	return t.Add(-e), t.Add(e), e
}

// wait obtains the current time and sleeps until future timestamps are guaranteed
// to be greater than the current time.
func wait() (t time.Time, e time.Duration) {
	t, e = time.Now(), epsilon()
	time.Sleep(e)
	return
}

func handleNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	earliest, latest, epsilon := ttNow()
	w.Header().Add(COMMIT.EarliestHeader, fmt.Sprintf("%d", earliest.UnixNano()))
	w.Header().Add(COMMIT.LatestHeader, fmt.Sprintf("%d", latest.UnixNano()))
	w.Header().Add(COMMIT.EpsilonDebugHeader, fmt.Sprintf("%d", epsilon.Nanoseconds()))
	// TODO: signature
}

func handleCommitWait(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	t, e := wait()
	w.Header().Add(COMMIT.EpsilonDebugHeader, fmt.Sprintf("%d", e.Nanoseconds()))
	w.Header().Add(COMMIT.TimestampHeader, fmt.Sprintf("%d", t.UnixNano()))
	// TODO: signature
}

func main() {
	flag.Parse()
	m := http.NewServeMux()
	m.HandleFunc("/.well-known/truetime/now", handleNow)
	m.HandleFunc("/.well-known/truetime/commitwait", handleCommitWait)
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: m,
		// TODO: TLSConfig
	}
	log.Fatal(s.ListenAndServe())
}
