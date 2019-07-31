// faketime serves fake TrueTime intervals with some random
// jitter added around the system clock.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var port = flag.Int("port", 8080, "port to listen on (default 8080)")
var jitterMs = flag.Float64("jitter", 10.0, "stddev of timestamp range in milliseconds (default 10.0)")

type interval struct {
	earliest, latest int64
}

// epislon returns a random duration drawn from a normal distribution with
// mean 0 and standard deviation *jitterMs (in milliseconds)
func epsilon() time.Duration {
	return time.Duration(rand.NormFloat64() * *jitterMs)
}

func ttNow() (earliest, latest time.Time) {
	t := time.Now()
	e := epsilon()
	return t.Add(e), t.Add(-e)
}

func wait() time.Time {
	t := time.Now()
	time.Sleep(epsilon() * time.Millisecond)
	return t
}

func handleNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	earliest, latest := ttNow()
	w.Header().Add("Consistent-Range-Earliest", fmt.Sprintf("%d", earliest.UnixNano()/1000))
	w.Header().Add("Consistent-Range-Latest", fmt.Sprintf("%d", latest.UnixNano()/1000))
	// TODO: signature
}

func handleCommitWait(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Consistent-Timestamp", fmt.Sprintf("%d", wait().UnixNano()/1000))
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
