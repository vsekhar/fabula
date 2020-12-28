package peerbook_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/vsekhar/fabula/internal/peerbook"
)

const firstControlPort = 28679
const firstTagPort = 29679

func createPeers(t *testing.T, n int) (peerbooks []*peerbook.PeerBook) {
	peerbooks = make([]*peerbook.PeerBook, n)
	for i := 0; i < n; i++ {
		peerbooks[i] = peerbook.New(
			fmt.Sprintf("peerbook_test_peer_id_%d", i),
			firstControlPort+i,
			map[string]string{
				"port": fmt.Sprintf("%d", firstTagPort+i),
			},
		)
	}
	return peerbooks
}

func startPeers(t *testing.T, ctx context.Context, pbs []*peerbook.PeerBook) {
	for _, p := range pbs {
		if err := p.Start(ctx); err != nil {
			t.Error(err)
		}
	}
}

func linkLocalPeers(t *testing.T, pbs []*peerbook.PeerBook) {
	if len(pbs) < 2 {
		return
	}
	for i := 1; i < len(pbs); i++ {
		pb := pbs[i]
		n, err := pb.Join([]string{fmt.Sprintf("localhost:%d", pbs[i-1].Port())})
		if err != nil {
			t.Fatal(err)
		}
		if n != 2 {
			t.Errorf("expected to join 2, got %d", n)
		}
	}
}

func waitForShutdowns(t *testing.T, pbs []*peerbook.PeerBook) {
	wg := new(sync.WaitGroup)
	wg.Add(len(pbs))
	for _, p := range pbs {
		go func(p *peerbook.PeerBook) {
			p.WaitForShutdown()
			wg.Done()
		}(p)
	}
	wg.Wait()
}

func TestPeerBook(t *testing.T) {
	peers := createPeers(t, 2)
	defer waitForShutdowns(t, peers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startPeers(t, ctx, peers)
	linkLocalPeers(t, peers)
	if p := peers[1].PeerCount(); p != 2 {
		t.Errorf("expected 2 peer, got %d peers", p)
	}
}

const httpServerString = "peerbook_test"

func httpServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(httpServerString))
	})
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}

type peerObject struct {
	hclient *http.Client
	port    int
}

func TestPeerObject(t *testing.T) {
	peerCount := 2
	peers := createPeers(t, peerCount)
	defer waitForShutdowns(t, peers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svrs := make([]*http.Server, peerCount)

	for i, p := range peers {
		// Client
		p.NewPeerObject = func(name string, addr net.IP, tags map[string]string) (interface{}, error) {
			portString, ok := tags["port"]
			if !ok {
				t.Fatal("peer did not provide a port in tags")
			}
			p, err := strconv.Atoi(portString)
			if err != nil {
				return nil, err
			}
			return &peerObject{
				hclient: &http.Client{},
				port:    p,
			}, nil
		}

		// Server
		svrs[i] = httpServer(firstTagPort + i)
		go svrs[i].ListenAndServe()
	}
	defer func() {
		for _, svr := range svrs {
			svr.Shutdown(context.Background())
		}
	}()

	startPeers(t, ctx, peers)
	linkLocalPeers(t, peers)
	obj, err := peers[0].GetPeerObject("a127")
	if err != nil {
		t.Fatal(err)
	}
	pobj := obj.(*peerObject)
	resp, err := pobj.hclient.Get(fmt.Sprintf("http://localhost:%d/", pobj.port))
	if err != nil {
		t.Fatal(err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(respBytes, []byte(httpServerString)) {
		t.Errorf("expected %s, got %s", []byte(httpServerString), respBytes)
	}
}

type broadcastHandler struct {
	wg *sync.WaitGroup
	t  *testing.T
}

func (b *broadcastHandler) Handle(ltime peerbook.LamportTime, name string, payload []byte, coalesce bool) {
	b.t.Log("broadcast: ", ltime, name, string(payload), coalesce)
	b.wg.Done()
}

func TestBroadcast(t *testing.T) {
	peerCount := 2
	peers := createPeers(t, peerCount)
	defer waitForShutdowns(t, peers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handlers := make([]*broadcastHandler, peerCount)

	for i, p := range peers {
		handlers[i] = &broadcastHandler{wg: new(sync.WaitGroup), t: t}
		handlers[i].wg.Add(1)
		p.BroadcastHandler = handlers[i].Handle
	}
	startPeers(t, ctx, peers)
	linkLocalPeers(t, peers)

	// We cannot easily test coalescing behavior since it relies on asynchronous
	// congestion among peers.

	if err := peers[0].Broadcast("peerbook_test_broadcast_name", []byte("peerbook_test_broadcast_payload"), false); err != nil {
		t.Error(err)
	}
	for _, h := range handlers {
		h.wg.Wait()
	}
}
