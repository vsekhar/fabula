package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var port = flag.Int("port", 0, "port to listen on (default system assigned)")
var timeServer = flag.String("timeserver", "localhost:8080", "URL of time server (default localhost:8080)")

type commitServer struct {
	data map[string]string
}

func (c *commitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle PUT (with consistency ID)

	// TODO: Handle COMMIT
}

func main() {
	flag.Parse()
	cs := new(commitServer)
	cs.data = make(map[string]string)
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: cs,
	}
	log.Fatal(s.ListenAndServe())
}
