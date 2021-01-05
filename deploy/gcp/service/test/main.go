package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/vsekhar/fabula/deploy/gcp/service/test/hellopb"
)

var port = flag.Int("port", 0, "port to listen on")

type server struct {
	hellopb.UnimplementedHelloServer
}

func (*server) Hello(ctx context.Context, r *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	return &hellopb.HelloResponse{Message: "Hello"}, nil
}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port %s", port)
	srv := grpc.NewServer()
	hellopb.RegisterHelloServer(srv, &server{})
	if err := srv.Serve(l); err != nil {
		log.Fatal(err)
	}
}
