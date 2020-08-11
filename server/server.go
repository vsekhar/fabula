package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/spanner"
)

type server struct {
	spclient *spanner.Client
	region   string
}

func newServer(db string) *server {
	ret := &server{}

	// https://cloud.google.com/compute/docs/storing-retrieving-metadata#querying
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/zone", nil)
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("error getting instance metadata:", err)
	}
	b := new(bytes.Buffer)
	io.Copy(b, resp.Body)
	metadata := b.String()
	// e.g.: "projects/469915952228/zones/us-central1-1"
	// On Cloud Run, the last "-1" is a fixed suffix
	parts := strings.Split(metadata, "/")
	if len(parts) != 4 {
		log.Fatal("unexpected metadata string:", metadata)
	}
	ret.region = parts[3][:len(parts[3])-2]

	bg := context.Background()
	ret.spclient, err = spanner.NewClient(bg, db)
	if err != nil {
		log.Fatal("error creating Spanner client:", err)
	}
	return ret
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello world, from %s.", s.region)
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s, from %s!\n", target, s.region)
}
