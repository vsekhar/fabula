// faketime serves fake TrueTime intervals with some random
// jitter added around the system clock.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/vsekhar/COMMIT/pkg/commit"
	"github.com/vsekhar/COMMIT/pkg/faketime"
)

var port = flag.Int("port", 8080, "port to listen on (default 8080)")
var jitterMs = flag.Float64("jitter", 10.0, "stddev of timestamp range in milliseconds (default 10.0)")

func main() {
	flag.Parse()
	m := http.NewServeMux()
	m.HandleFunc(commit.TrueTimeNowPath, faketime.NowHandler(*jitterMs))
	m.HandleFunc(commit.TrueTimeCommitWaitPath, faketime.CommitWaitHandler((*jitterMs)))
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: m,
		// TODO: TLSConfig
	}
	log.Fatal(s.ListenAndServe())
}
