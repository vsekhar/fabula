package main

import (
	"flag"
	"log"
	"net/http"
)

var timeServer = flag.String("timeserver", "", "address of timeserver (e.g. 'timeserver.com'")
var server1 = flag.String("server1", "", "address of first server (e.g. 'dataserver1.com'")
var server2 = flag.String("server2", "", "address of second server (e.g. 'dataserver2.com'")

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags)
	if *server1 == "" {
		log.Fatal("No server1 specified")
	}
	if *server2 == "" {
		log.Fatal("No server2 specified")
	}
	_, err := http.Get("http://" + *server1)
	if err != nil {
		log.Fatal(err)
	}
}
