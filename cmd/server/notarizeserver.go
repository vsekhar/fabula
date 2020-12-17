package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/vsekhar/fabula/pkg/api/servicepb"
	pb "github.com/vsekhar/fabula/pkg/api/servicepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: notarize server accepts requests from the public and forwards to the
// pack service

type notarizeServer struct {
	*http.ServeMux
	agent *agent.Agent

	servicepb.UnimplementedFabulaServer
}

func newNotarizeServer(name string, a *agent.Agent) *notarizeServer {
	mux := http.NewServeMux()

	// TODO: url format

	// TODO: view handlers: packs, entries, proofs

	// notarization handlers
	// TODO: POST handler for new notarization
	//  - do notarization
	//  - redirect to canonical URL for response
	mux.HandleFunc("/v1/notarize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Add("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// TODO: notarization_sha3512 = hash(document, salt)
		// TODO: lookup pack service for prefix = notarization_sha3512[:prefixLen]
		// TODO: submit to Pack Service, block until done
		// TODO: if fail, return error
		// TODO: if success, get timestamp, finish commit wait, return success
		// TODO: response includes: salt, notarization, prefix, timestamp, info
		//       about packs, debug info (commit-wait length), canonical URL
		// TODO: if redirect=true, redirect to canonical URL for response
		//       perhaps using HTML5 window.history.replaceState()
		fmt.Fprintf(w, "Hello world, from %s\n", name)
	})

	// system handlers
	mux.HandleFunc("/v1/system/peers", func(w http.ResponseWriter, r *http.Request) {
		members := a.Serf().Members()
		for _, m := range members {
			fmt.Fprintf(w, "%+v\n", m)
		}
	})

	// liveness probe
	mux.HandleFunc("/_liveness", func(w http.ResponseWriter, r *http.Request) {
		// TODO: check for liveness
		return
	})

	return &notarizeServer{
		ServeMux: mux,
		agent:    a,
	}
}

func (s *notarizeServer) Notarize(ctx context.Context, r *pb.NotarizeRequest) (*pb.NotarizeResponse, error) {
	// TODO: notarization_sha3512 = hash(prior, document, timestamp)
	// TODO: submit to Pack Service, block until done
	// TODO: if fail, return error
	// TODO: if success, get timestamp, finish commit wait, return success
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *notarizeServer) Root(ctx context.Context, r *pb.RootRequest) (*pb.RootResponse, error) {
	return &pb.RootResponse{Message: "Root"}, nil
}
