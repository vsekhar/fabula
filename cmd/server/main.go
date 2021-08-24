package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache/consistenthash"
	"github.com/gorilla/handlers"
	lru "github.com/hashicorp/golang-lru"
	"github.com/kenshaw/sdhook"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"

	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"

	"cloud.google.com/go/storage"

	internalapi "github.com/vsekhar/fabula/internal/api"
	"github.com/vsekhar/fabula/internal/interrupt"
	"github.com/vsekhar/fabula/pkg/api/servicepb"
)

var (
	port            = flag.Int("port", 0, "port for web server (default: auto)")
	notarizeRPCPort = flag.Int("notarizerpcport", 0, "rpc port for notarization (default: auto)")
	packRPCPort     = flag.Int("packrpcport", 0, "rpc port for packing (default: auto)")
	controlPort     = flag.Int("controlport", 7946, "rpc port for P2P cluster control")
	bucketName      = flag.String("bucket", "", "bucket to store sequence to (e.g. 'gcs://bucket_name')")
	join            = flag.String("join", "", "internal host:port of other servers to join with")
	userEventPeriod = flag.Duration("usereventperiod", time.Duration(0), "period with which to send a user event")
	verbose         = flag.Bool("verbose", false, "verbose log level")
	dev             = flag.Bool("dev", false, "dev mode (human-readable) logging")
)

// The number of shards to divide the keyspace into for each server in the
// cluster. A higher number will ensure a more evently divided keyspace for
// smaller numbers of servers. There is no real cost to having a very high
// number since keys are discrete and independent: we don't gain anything by
// having nearby keys in the keyspace allocated to the same server, only that
// individual keys are consistently allocated to the same server on each client.
const shardsPerServer = 100

const memberTimer = 10 * time.Second

const role = "fabula-server"

type ringMux struct {
	memberFn func() []serf.Member
	ring     atomic.Value  // *consistenthash.Map[string(prefix)]string(name)
	members  *sync.Map     // map[string(name)]serf.Member
	clients  *lru.ARCCache // effectively map[string(name)]*client
}

// safe to call from multiple goroutines
func (r *ringMux) updateMembers() {
	members := r.memberFn()
	alive := 0
	for _, m := range members {
		if m.Status == serf.StatusAlive {
			alive++
		}
	}

	nameMap := make(map[string]struct{})
	addrMap := make(map[string]struct{})
	for _, m := range members {
		if m.Tags["role"] == role &&
			m.Status == serf.StatusAlive {
			if m.Name == "" || m.Addr.String() == "" {
				continue
			}
			addrMap[m.Addr.String()] = struct{}{}
			nameMap[m.Name] = struct{}{}
			r.members.Store(m.Name, m)
		}
	}
	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		log.Printf("[DEBUG] found member: %s", name)
		names = append(names, name)
	}

	// package consistenthash refers to shardsPerServer as "replicas" which
	// isn't really accurate since there is no replication involved.
	newRing := consistenthash.New(shardsPerServer /* "replicas" */, nil)
	newRing.Add(names...)
	r.ring.Store(newRing)

	// Prevent members map from growing endlessley.
	r.members.Range(func(key, value interface{}) bool {
		name := key.(string)
		if _, ok := nameMap[name]; !ok {
			r.members.Delete(key)
			log.Printf("[DEBUG] dropping member: %s", name)
		}
		return true // continue with range call
	})
	log.Printf("[DEBUG] main: hashring size: %d - %#v", len(addrMap), addrMap)
}

// TODO: redirecting via ringMux

// Handle a serf.Event.
//
// Implements interface agent.EventHandler.
func (r *ringMux) HandleEvent(e serf.Event) {
	switch x := e.(type) {
	case serf.MemberEvent:
		r.updateMembers()
	case serf.UserEvent:
		log.Printf("[INFO] User event received: %+v", x)
	}
}

func main() {
	flag.Parse()

	if *dev {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetReportCaller(true)
	h, err := sdhook.New(
		sdhook.GoogleLoggingAgent(),
		sdhook.ErrorReportingService("fabula"),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.AddHook(h)

	/*
		// OLD hclog logging
		// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.SetFlags(log.Lshortfile)
		logOpts := &hclog.LoggerOptions{
			Name:            "fabula-server",
			IncludeLocation: true,
		}
		if *verbose {
			logOpts.Level = hclog.Level(hclog.Debug)
		}
		if *dev {
			logOpts.Color = hclog.AutoColor
		} else {
			logOpts.JSONFormat = true
		}
		appLogger := hclog.New(logOpts)
		logWriter := appLogger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})
		stdLogger := appLogger.StandardLogger(&hclog.StandardLoggerOptions{InferLevels: true})
		log.SetOutput(logWriter)
		log.SetPrefix("")
		log.SetFlags(0)
	*/

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	name, err := getPodName()
	if err != nil {
		log.WithError(err).Fatal("could not get pod name")
	}
	log.Printf("instance name: %s", name)

	if *bucketName == "" {
		log.Fatalf("[ERROR] -bucket required")
	}

	// TODO: fix this to get the pod hostname

	var a *agent.Agent // forward declare for handlers

	// Web service
	weblistener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("main: opening web listen port %d: %s", *port, err)
	}

	notarizeSvr := newNotarizeServer(name, a)
	websrv := &http.Server{
		Addr:    weblistener.Addr().String(),
		Handler: handlers.LoggingHandler(os.Stdout, notarizeSvr),
	}
	go websrv.Serve(weblistener)
	defer websrv.Shutdown(ctx)
	log.Printf("[INFO] main: web server listening at %s", websrv.Addr)

	_, webListenerPort, err := net.SplitHostPort(weblistener.Addr().String())
	if err != nil {
		log.Fatalf("[ERROR] main: splitting web addr:port: %s", err)
	}

	// RPC notarize service
	rpcNotarizeListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *notarizeRPCPort))
	if err != nil {
		log.Fatalf("[ERROR] main: opening notarize rpc listen port %d: %s", *notarizeRPCPort, err)
	}
	_ = rpcNotarizeListener
	notarizerpcsrv := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	notarizesvr := newNotarizeServer(name, a)
	servicepb.RegisterFabulaServer(notarizerpcsrv, notarizesvr)
	go notarizerpcsrv.Serve(rpcNotarizeListener)
	defer notarizerpcsrv.Stop()
	log.Printf("[INFO] main: notarize rpc server listening at %s", rpcNotarizeListener.Addr())

	_, notarizeRPCListenerPort, err := net.SplitHostPort(rpcNotarizeListener.Addr().String())
	if err != nil {
		log.Fatalf("[ERROR] main: splitting addr:port: %w", err)
	}

	// Pack service storage
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("[ERROR] main: creating storage client: %s", err)
	}
	cred, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		log.Fatalf("[ERROR] main: getting default credentials: %s", err)
	}
	bkt := client.Bucket(*bucketName).UserProject(cred.ProjectID)

	// RPC pack service
	rpclistener, err := net.Listen("tcp", fmt.Sprintf(":%d", *packRPCPort))
	if err != nil {
		log.Fatalf("[ERROR] main: opening pack rpc listen port %d: %s", *packRPCPort, err)
	}
	packrpcsrv := grpc.NewServer()
	packsvr := newPackServer(ctx, bkt)
	internalapi.RegisterPackerServer(packrpcsrv, packsvr)
	go packrpcsrv.Serve(rpclistener)
	defer packrpcsrv.Stop()
	log.Printf("[INFO] main: pack rpc server listening at %s", rpclistener.Addr())

	_, packRPCListenerPort, err := net.SplitHostPort(rpclistener.Addr().String())
	if err != nil {
		log.Fatalf("[ERROR] main: splitting addr:port: %w", err)
	}

	// Set up internal gossip network
	log.Printf("[INFO] main: using control port: %d", *controlPort)
	serfConfig := serf.DefaultConfig()
	serfConfig.NodeName = name
	serfConfig.MemberlistConfig.BindAddr = "0.0.0.0"
	serfConfig.MemberlistConfig.BindPort = *controlPort
	// serfConfig.Logger = stdLogger
	serfConfig.Tags = map[string]string{
		"role":                     role,
		"fabula-notarize-web-port": webListenerPort,
		"fabula-notarize-rpc-port": notarizeRPCListenerPort,
		"fabula-pack-rpc-port":     packRPCListenerPort,
	}
	agentConfig := agent.DefaultConfig()
	agentConfig.Discover = "serf.server.fabula-2020-12-14.svc.cluster.local"
	agentConfig.NodeName = name
	agentConfig.BindAddr = fmt.Sprintf("0.0.0.0:%d", *controlPort) // seems unused
	agentConfig.LogLevel = "ERROR"                                 // seems unused
	// agentConfig.Tags seems unused
	a, err = agent.Create(agentConfig, serfConfig, os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Shutdown()
	defer a.Leave()

	// Set up ring muxer
	rm := &ringMux{
		memberFn: func() []serf.Member { return a.Serf().Members() },
		members:  new(sync.Map),
	}
	a.RegisterEventHandler(rm)
	go func() {
		for range time.NewTicker(memberTimer).C {
			rm.updateMembers()
		}
	}()

	if err := a.Start(); err != nil {
		log.Fatal(err)
	}
	local := a.Serf().LocalMember()
	log.Printf("listening for internal connections at: %s:%d", local.Addr.String(), local.Port)
	remotes := strings.Split(*join, " ")
	if *join != "" && len(remotes) > 0 {
		log.Printf("joining %d remotes: %v", len(remotes), remotes)
		n, err := a.Join(remotes, false)
		if err != nil {
			log.Fatal(err)
		}
		if n != len(remotes)+1 {
			// (we join ourselves too)
			log.Fatalf("joined %d, expected %d", n, len(remotes))
		}
	}

	if *userEventPeriod > time.Duration(0) {
		go func() {
			for t := range time.Tick(*userEventPeriod) {
				a.UserEvent(fmt.Sprintf("event @ %s", t), nil, true)
			}
		}()
	}

	if err := populateSerfFromK8s(ctx, a); err != nil {
		if err != errNotInCluster {
			log.Fatal(err)
		}
	}

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
}
