package main

import (
	"context"
	_ "expvar"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/vsekhar/fabula/deploy/gcp/service/test/hellopb"
	"github.com/vsekhar/fabula/internal/cloudenv"
)

var port = flag.Int("port", 0, "port to listen on")
var downstream = flag.String("downstream", "", "hostname of downstream service; used for testing an internal service via an external one")

type server struct {
	hostname         string
	downstreamClient hellopb.HelloClient
	hellopb.UnimplementedHelloServer
}

func (s *server) Hello(ctx context.Context, r *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	var downstreamResp *hellopb.HelloResponse
	if *downstream != "" {
		var err error
		downstreamResp, err = s.downstreamClient.Hello(ctx, &hellopb.HelloRequest{})
		if err != nil {
			log.WithContext(ctx).WithField("hostname", s.hostname).WithError(err).Error("fetching downstream")
			return nil, err
		}
	}
	return &hellopb.HelloResponse{
		Message:    "Hello",
		Hostname:   s.hostname,
		Timestamp:  ptypes.TimestampNow(),
		Downstream: downstreamResp,
	}, nil
}

func main() {
	flag.Parse()
	ctx, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()

	c, err := cloudenv.Get(ctx, "svc_test", false, false)
	if err != nil {
		log.WithError(err).Fatal("getting cloudenv")
	}

	logFields := make(log.Fields)
	logFields["hostname"] = c.Hostname()
	logFields["cloudenv"] = c.CloudEnv()
	go func() {
		for {
			p, err := c.Peers(ctx, "")
			if err != nil {
				log.WithFields(logFields).WithError(err).Fatal("getting peers")
			}
			log.WithFields(log.Fields{
				"peers":    p,
				"cloudenv": c.CloudEnv(),
			}).Info("fetched peers")
			time.Sleep(3 * time.Second)
		}
	}()
	logFields["port_arg"] = *port
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.WithFields(logFields).WithError(err).Fatal("opening listen port")
	}
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		log.WithFields(logFields).WithError(err).Fatal("splitting listen host and port")
	}
	logFields["port"] = port
	log.WithFields(logFields).Printf("listening")
	srv := grpc.NewServer()
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(logFields).WithError(err).Fatal("getting hostname")
	}
	logFields["hostname"] = hostname
	var downstreamClient hellopb.HelloClient
	if *downstream != "" {
		logFields["downstream"] = *downstream
		conn, err := grpc.Dial(*downstream, grpc.WithInsecure())
		if err != nil {
			log.WithFields(logFields).WithError(err).Fatalf("dialing downstream")
		}
		downstreamClient = hellopb.NewHelloClient(conn)
	}
	hellosrv := &server{
		hostname:         hostname,
		downstreamClient: downstreamClient,
	}
	hellopb.RegisterHelloServer(srv, hellosrv)
	if err := srv.Serve(l); err != nil {
		log.WithFields(logFields).WithError(err).Fatal("serve returned")
	}
}
