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

type fakeTime struct{}

func (*fakeTime) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.URL.Path != "/.well-known/truetime/now" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	midPointMs := time.Now().UnixNano() / 1000
	epsilon := int64(rand.NormFloat64() * *jitterMs)
	earliest := midPointMs - epsilon
	latest := midPointMs + epsilon
	h := w.Header()
	h.Add("Consistent-Range-Earliest", fmt.Sprintf("%d", earliest))
	h.Add("Consistent-Range-Latest", fmt.Sprintf("%d", latest))
	// TODO: signature
}

func main() {
	flag.Parse()
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: new(fakeTime),
	}
	log.Fatal(s.ListenAndServe())
}
