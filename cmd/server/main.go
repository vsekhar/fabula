package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"

	batcherpb "github.com/vsekhar/fabula/internal/api/batcher"
	"github.com/vsekhar/fabula/internal/atomicwriter"
	"github.com/vsekhar/fabula/internal/initlogrus"
	"github.com/vsekhar/fabula/internal/interrupt"
	"github.com/vsekhar/fabula/internal/peerbook"
)

var (
	port        = flag.Int("port", 0, "port for web server (default: auto)")
	controlPort = flag.Int("controlport", 7946, "rpc port for cluster P2P")
	groupName   = flag.String("groupname", "", "name of service or instance group to poll for peers")

	storagePath = flag.String("storage", "", "path to store sequence and packs (e.g. '.' or 'gcs://bucket_name')")
	// TODO: verifyTopic
	// TODO: sequenceTopic (for the public)

	verbose = flag.Bool("verbose", false, "verbose log level")
	dev     = flag.Bool("dev", false, "dev mode (human-readable) logging")
)

const role = "fabula-server"

type joinValue []string

func (j *joinValue) String() string {
	if j == nil {
		return ""
	}
	return strings.Join(*j, ",")
}

func (j *joinValue) Set(s string) error {
	*j = append(*j, s)
	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	joinFlags := new(joinValue)
	flag.Var(joinFlags, "join", "internal host:port of other servers to join with")
	flag.Parse()
	initlogrus.Init(*dev, *verbose)

	ctx, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()

	name, ok := os.LookupEnv("FABULA_INSTANCE_NAME")
	if !ok {
		// Try k8s
		var err error
		name, err = getPodName()
		if err != nil {
			if err != errNotInCluster {
				log.WithError(err).Fatal("could not get pod name")
			}
			name = fmt.Sprintf("fabula-local-%d", rand.Intn(1<<16))
		}

		// TODO: try GCE mig

		// TODO: while at it, get a function to return the IPs of other nodes
		// in the group and call peerbook.Join on it later.
	}
	log.WithField("name", name).Info("starting")

	closeTelemetry := initTelemetry(ctx)
	defer closeTelemetry()

	if *storagePath == "" {
		log.Fatalf("[ERROR] -storage required")
	}

	// All nodes should create a sequencer object ready to take over sequencing.
	// It will listen for broadcasts from PeerBook to maintain a high-water-mark
	// and handle batching and writing (via atomicwriter) of sequenced entries.
	awd, err := atomicwriter.NewDriver(ctx, *storagePath)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"path": *storagePath,
		}).Fatal("could not write to path")
	}
	_ = awd
	// TODO: implement

	// Peer-to-peer batching service
	internalListener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.WithError(err).Fatal("opening ephemeral listen port")
	}
	_, internalBatcherPort, err := net.SplitHostPort(internalListener.Addr().String())
	if err != nil {
		log.WithError(err).Fatal("getting ephemeral listen port")
	}
	log.WithField("rpcPort", internalBatcherPort).Info("internal ephemeral batcher listen port")
	peerCtx, cancelPeers := context.WithCancel(ctx)
	peers, err := peerbook.New(peerCtx, name, *controlPort, map[string]string{
		"servicePort": internalBatcherPort,
	})
	if err != nil {
		log.WithError(err).Fatal("creating peerbook")
	}

	internalSvr := newInternalBatcherServer(peers)
	internalGRPCSvr := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	batcherpb.RegisterInternalBatcherServer(internalGRPCSvr, internalSvr)
	go internalGRPCSvr.Serve(internalListener)
	defer internalGRPCSvr.GracefulStop()
	defer peers.WaitForShutdown()
	defer cancelPeers()
	log.WithField("controlPort", peers.Port()).Info("peer control port")

	// Join peers provided on command line
	if len(*joinFlags) > 0 {
		n, err := internalSvr.peers.Join(*joinFlags)
		if err != nil {
			log.WithError(err).WithField("joins", *joinFlags).Errorf("joining peers from flags")
		}
		// TODO: confirm that there is always just one extra (ourselves?)
		if n != len(*joinFlags)+1 {
			log.WithFields(log.Fields{
				"expected": len(*joinFlags),
				"actual":   n,
			}).Error("joins")
		}
	}

	// TODO: Periodically seed peer group from k8s or GCE MIG
	if metadata.OnGCE() {

	}
	/*
		if err := populateSerfFromK8s(ctx, a); err != nil {
			if err != errNotInCluster {
				log.Fatal(err)
			}
		}
	*/

	// TODO: public GRPC service
	// TODO: web service that forwards to local public GRPC service

	// TODO: create new short-lived subscriber, use pkg/latest to track HWM, use HWM
	// when hashing new entries, submit new entries to pubsub.

	// TODO: create a logWriter struct that tracks the latest sequence number
	// for a prefix, bundles new values and writes them to the log.

	// TODO: put logWriters in an LRU cache. Individual servers don't know/care
	// which prefixes they own. They trust other servers to use ringMux
	// correctly. In other words, from a server's point of view, if a request
	// arrives, it is correct and should be handled.

	// TODO: put prefixes and latest sequence numbers in groupcache, so servers
	// have something to go by when starting to handle a new prefix.
	// #optimization

	// TODO: handler for writing notarizing, forwarding based on hash ring

	interrupt.Wait()
	log.Info("Interrupt received, exiting")
}
