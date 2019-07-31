package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var port = flag.Int("port", 0, "port to listen on (default system assigned)")
var timeServer = flag.String("timeserver", "localhost:8080", "URL of time server (default localhost:8080)")

type value struct {
	data      string
	timestamp time.Time
}

type multiVersionValue []value

type commitServer struct {
	data map[string]multiVersionValue
	// TODO: use sync.Map
}

func (c *commitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle PUT (with consistency ID)

	// TODO: Handle COMMIT
}

func newCommitServer() *commitServer {
	cs := new(commitServer)
	cs.data = make(map[string]multiVersionValue)
	return cs
}

func main() {
	flag.Parse()
	cs := new(commitServer)
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: cs,
	}
	log.Fatal(s.ListenAndServe())
}
