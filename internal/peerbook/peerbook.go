// Package peerbook implements a continuously updated eventually consistent
// phonebook of peers.
package peerbook

import (
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/golang/groupcache/consistenthash"
	lru "github.com/hashicorp/golang-lru"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
	log "github.com/sirupsen/logrus"
	"github.com/vsekhar/fabula/internal/notify"
)

// The number of shards to divide the keyspace into for each server in the
// cluster. A higher number will ensure a more evently divided keyspace for
// smaller numbers of servers. There is no real cost to having a very high
// number since keys are discrete and independent: we don't gain anything by
// having nearby keys in the keyspace allocated to the same server, only that
// individual keys are consistently allocated to the same server on each client.
//
// For very large numbers of servers, many shards per server increases overhead
// (e.g. 500 servers x 100 shards means we must build a table of 50,000 shards
// each time a server is added or removed). But we do this asynchronously so it
// should be manageable.
const shardsPerServer = 100

type eventHandler struct {
	p *PeerBook
}

func (e *eventHandler) HandleEvent(event serf.Event) {
	switch x := event.(type) {
	case serf.MemberEvent:
		e.p.memberNotifier.Notify()
	case serf.UserEvent:
		log.Infof("User event received: %+v", x)
		if e.p.BroadcastHandler != nil {
			e.p.BroadcastHandler(
				x.LTime,
				x.Name,
				x.Payload,
				x.Coalesce,
			)
		}
	}
}

// LamportTime is a monotonic clock that can be used to order broadcast
// messages.
type LamportTime = serf.LamportTime

// PeerBook is an instance of a peer book.
//
// Exported fields shold be set before calling Start(), and should not be
// changed thereafter.
type PeerBook struct {
	port  int
	agent *agent.Agent

	memberNotifier *notify.Notifier
	ring           atomic.Value  // *consistenthash.Map[string(prefix)]string(name)
	members        *sync.Map     // map[string(name)]serf.Member
	clients        *lru.ARCCache // effectively map[string(name)]*client

	// BroadcastHandler is called to handle each broadcast message sent to
	// peers. It should complete quickly to prevent missing future messages.
	BroadcastHandler func(ltime LamportTime, name string, payload []byte, coalesce bool)

	// NewPeerObject is a function that returns an object corresponding to the
	// peer whose name, address and tags are provided. PeerBook will cache these
	// objects and provide them via GetPeerObject().
	//
	// Peer objects are usually used to cache client objects or connection
	// objects used to contact peers.
	NewPeerObject func(name string, addr net.IP, tags map[string]string) (interface{}, error)

	// DestroyPeerObject is called when a peer object is no longer useful (e.g.
	// when the peer disappears).
	//
	// DestroyPeerObject should complete quickly. Ohterwise, PeerBook may fail
	// to keep up with changes in membership among peers.
	//
	// If DestroyPeerObject is nil, then the peer object is simply dropped. If
	// PeerBook held the only reference to the peer object, then the object will
	// be deallocated.
	DestroyPeerObject func(key string, obj interface{})
	peerObjects       *sync.Map
}

// New returns a new PeerBook
func New(nodeName string, port int, tags map[string]string) *PeerBook {
	serfConfig := serf.DefaultConfig()
	serfConfig.NodeName = nodeName
	serfConfig.MemberlistConfig.BindAddr = "0.0.0.0"
	serfConfig.MemberlistConfig.BindPort = port
	serfConfig.MemberlistConfig.LogOutput = os.Stdout
	// serfConfig.Logger = stdLogger
	serfConfig.LogOutput = os.Stdout
	serfConfig.Tags = tags
	agentConfig := agent.DefaultConfig()
	agentConfig.NodeName = nodeName

	// These seem unused:
	//   agentConfig.BindAddr = fmt.Sprintf("0.0.0.0:%d", port)
	//   agentConfig.LogLevel = "ERROR"
	//   agentConfig.Tags = tags

	// Doesn't work on GCP, no broadcast:
	//   agentConfig.Discover = "peerbook.svc.cluster.local"

	agent, err := agent.Create(agentConfig, serfConfig, os.Stderr)
	if err != nil {
		log.WithError(err).Fatal("peerbook: should not occur, agent.Create only returns error if loading a tags or keyring file fails, we use neither")
	}

	r := new(PeerBook)
	r.port = port
	r.agent = agent
	r.members = new(sync.Map)
	r.memberNotifier = notify.New(r.updateMembers)
	r.peerObjects = new(sync.Map)
	agent.RegisterEventHandler(&eventHandler{p: r})
	return r
}

// Start starts the PeerBook. Exported fields should not be changed after
// calling start.
func (p *PeerBook) Start() error {
	return p.agent.Start()
}

func (p *PeerBook) updateMembers() {
	members := p.agent.Serf().Members()
	alive := 0
	for _, m := range members {
		if m.Status == serf.StatusAlive {
			alive++
		}
	}

	nameMap := make(map[string]struct{})
	addrMap := make(map[string]struct{})
	for _, m := range members {
		if m.Status == serf.StatusAlive {
			if m.Name == "" || m.Addr.String() == "" {
				continue
			}
			nameMap[m.Name] = struct{}{}
			addrMap[m.Addr.String()] = struct{}{}
			p.members.Store(m.Name, m)
		}
	}
	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}

	// TODO: maglev hash?

	// package consistenthash refers to shardsPerServer as "replicas" which
	// isn't really accurate since there is no replication involved.
	newRing := consistenthash.New(shardsPerServer /* "replicas" */, nil)
	newRing.Add(names...)
	p.ring.Store(newRing)

	// Prevent member and peerobject maps from growing endlessly.
	p.members.Range(func(key, value interface{}) bool {
		name := key.(string)
		if _, ok := nameMap[name]; !ok {
			p.members.Delete(key)
			log.Debugf("peerbook: dropping member: %s", name)
		}
		return true // keep going
	})
	p.peerObjects.Range(func(key, value interface{}) bool {
		name := key.(string)
		if _, ok := nameMap[name]; !ok {
			if obj, loaded := p.peerObjects.LoadAndDelete(key); loaded && p.DestroyPeerObject != nil {
				p.DestroyPeerObject(name, obj)
			}
			log.Debugf("peerbook: purging peerObject for member: %s", name)
		}
		return true // keep going
	})

	log.Debugf("peerbook: hashring size: %d - %#v", len(addrMap), addrMap)
	// TODO: report metric
}

// Close winds down a PeerBook and releases resources.
func (p *PeerBook) Close() error {
	err := p.agent.Leave()
	p.agent.Shutdown()
	p.memberNotifier.Close()
	if p.DestroyPeerObject != nil {
	}
	p.peerObjects.Range(func(key, value interface{}) bool {
		if p.DestroyPeerObject != nil {
			p.DestroyPeerObject(key.(string), value)
		}
		return true // keep going
	})
	return err
}

// Broadcast broadcasts a small message to all peers in the group. Broadcasts
// can be used to keep peers up-to-date. Broadcasts are delivered on an
// eventually consistent basis.
func (p *PeerBook) Broadcast(name string, payload []byte, coalesce bool) error {
	return p.agent.UserEvent(name, payload, coalesce)
}

// GetPeer returns the peerName, ip and tags of the peer owning the provided
// key.
func (p *PeerBook) GetPeer(key string) (peerName string, addr net.IP, tags map[string]string) {
	peerName = p.ring.Load().(*consistenthash.Map).Get(key)
	memberi, ok := p.members.Load(peerName)
	if !ok {
		log.Fatalf("peerbook: expected to load member information for: %s", peerName)
	}
	member := memberi.(serf.Member)
	return peerName, member.Addr, member.Tags
}

// GetPeerObject returns an object associated with the peer owning the provided
// key. Objects are created from the NewPeerObject function member in the
// PeerBook.
//
// Peer objects are typically used for client connections.
func (p *PeerBook) GetPeerObject(key string) (interface{}, error) {
	// Short-circuit
	obj, ok := p.peerObjects.Load(key)
	if ok {
		return obj, nil
	}

	if p.NewPeerObject == nil {
		return nil, nil
	}

	// Create or use old (in case we lose the race on p.peerObjects)
	peerName, addr, tags := p.GetPeer(key)
	newobj, err := p.NewPeerObject(peerName, addr, tags)
	if err != nil {
		return nil, err
	}
	obj, loaded := p.peerObjects.LoadOrStore(key, newobj)
	if loaded && p.DestroyPeerObject != nil {
		p.DestroyPeerObject(key, newobj)
	}
	return obj, nil
}

// PeerCount returns the number of alive peers.
func (p *PeerBook) PeerCount() int {
	n := 0
	members := p.agent.Serf().Members()
	for _, m := range members {
		if m.Status == serf.StatusAlive {
			n++
		}
	}
	return n
}

// Join joins peers at the addresses in addrs.
//
// Join must be used to bootstrap a new PeerBook by introducing it to another
// instance of PeerBook. Addresses need to be exchanged out of band (e.g.
// manually or by querying cloud infrastructure).
//
// Start must be called on this peer before calling Join.
func (p *PeerBook) Join(addrs []string) (n int, err error) {
	return p.agent.Serf().Join(addrs, true /* ignoreOld */)
}

// Port returns the port the PeerBook is using for control traffic.
func (p *PeerBook) Port() int {
	return p.port
}

// SetTags updates the tags associated with the current peer. The provided tags
// replace any previously set tags for the current peer. The new tags will be
// distributed to other peers on an eventually consistent basis.
//
// Tags are usually used to advertise information about services a peer provides
// to other peers, e.g. the port at which an RPC server can be reached.
func (p *PeerBook) SetTags(tags map[string]string) error {
	return p.agent.Serf().SetTags(tags)
}
