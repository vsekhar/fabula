package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var metadata string

func main() {
	db := os.Getenv("DB")
	if db == "" {
		log.Fatal("no DB environment variable")
	}

	// Protocol versions
	v1 := newServer(db)
	http.Handle ("/v1/", v1)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("no PORT environment variable")
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
} 
