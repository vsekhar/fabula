package main

import (
	"flag"
	"log"
)

var timeServer = flag.String("timeserver", "", "address:port of timeserver")
var server1 = flag.String("server1", "", "address:port of first server")
var server2 = flag.String("server2", "", "address:port of second server")

func main() {
	flag.Parse()
	if *server1 == "" {
		log.Fatal("No server1 specified")
	}
	if *server2 == "" {
		log.Fatal("No server2 specified")
	}
}
