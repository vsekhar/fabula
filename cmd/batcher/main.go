package main

import (
	"context"
	"flag"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	batcherpb "github.com/vsekhar/fabula/internal/api/batcher"
	"github.com/vsekhar/fabula/internal/atomicwriter"
	"github.com/vsekhar/fabula/internal/initlogrus"
	"github.com/vsekhar/fabula/internal/peerbook"
	"github.com/vsekhar/fabula/internal/telemetry"
)

const role = "fabula-batcher"

var (
	port        = flag.Int("port", 0, "port for rpc")
	controlPort = flag.Int("controlPort", 7496, "rpc port for cluster P2P")
	groupName   = flag.String("groupname", "", "name of service or instance group to poll for peers")

	pubsubTopic = flag.String("pubsubTopic", "", "pubsub topic to publish batches to")
	storagePath = flag.String("storage", "", "path to store sequence to (e.g. '.' or 'gcs://bucket_name/folder_name")

	dev     = flag.Bool("dev", false, "developer mode (human-readable log output)")
	verbose = flag.Bool("verbose", false, "verbose output")
)

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
	joinFlags := new(joinValue)
	flag.Var(joinFlags, "join", "internal host:port of other servers to join with")
	flag.Parse()
	initlogrus.Init("batcher", "v0.1.0", *dev, *verbose)
	log.Info("batcher starting...")

	ctx, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()
	closeTelemetry := telemetry.InitTelemetry(ctx)
	defer closeTelemetry()

	if *storagePath == "" {
		log.Fatal("-storage required")
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
	// TODO: implement sequencer

	// P2P batching RPC service
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

}
