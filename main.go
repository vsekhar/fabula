package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/spanner"
	"google.golang.org/appengine"
)

var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
var dbInstanceName = os.Getenv("DB_INSTANCE_NAME")
var dbDbName = os.Getenv("DB_DB_NAME")
var db = "projects/" + projectID + "/instances/" + dbInstanceName + "/databases/" + dbDbName

var client *spanner.Client
var clientOnce sync.Once

func clientSetup() {
	var err error
	client, err = spanner.NewClient(context.Background(), db)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/_ah/warmup", warmup)
	http.HandleFunc("/log", logHandler)
	http.HandleFunc("/", handler)
	appengine.Main()
}

func warmup(w http.ResponseWriter, r *http.Request) {
	clientOnce.Do(clientSetup)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	clientOnce.Do(clientSetup)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world")
}
