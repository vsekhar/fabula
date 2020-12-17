package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	pb "github.com/vsekhar/fabula/internal/api"
	"github.com/vsekhar/fabula/internal/bigarray"
	"github.com/vsekhar/fabula/internal/prefix"
	"github.com/vsekhar/fabula/pkg/autobundler"
	"github.com/vsekhar/fabula/pkg/sortablebase64"
	"golang.org/x/sync/singleflight"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxPackSize = 100

func packNamePrefix(prefix string) string {
	return fmt.Sprintf("%s-")
}

func packName(prefix string, seqNo int) string {
	return fmt.Sprintf("%s%s.pack", packNamePrefix(prefix), sortablebase64.EncodeUint64(uint64(seqNo)))
}

type prefixPacker struct {
	server        *packServer
	prefix        string // hex encoded
	lastTimestamp time.Time
	lastHash      []byte
	nextSeqNo     int
	bundler       *autobundler.AutoBundler
}

func newPrefixPacker(ctx context.Context, server *packServer, prefix string) (*prefixPacker, error) {
	r := &prefixPacker{
		server: server,
		prefix: prefix,
	}
	var dneErr error
	doesNotExist := func(i int) (atLastChecked bool, lastChecked int) {
		itr := server.bucket.Objects(ctx, &storage.Query{
			Prefix:      packNamePrefix(prefix),
			StartOffset: packName(prefix, i),
		})
		n := 0
		var lastObj *storage.ObjectAttrs
		var dneErr error
		for {
			var obj *storage.ObjectAttrs
			obj, dneErr = itr.Next()
			if dneErr == iterator.Done {
				break
			}
			if dneErr != nil {
				return true, i // to stop search, must check dneErr
			}
			n++
			lastObj = obj
			if itr.PageInfo().Remaining() == 0 {
				break // next call to itr.Next() will fetch, so stop here
			}
		}
		if n == 0 {
			return true, i
		}
		// TODO: parse lastObj.Name, verify prefix, get seqNo,
		// return true, seqNo+1
		_ = lastObj
		return false, i
	}
	r.nextSeqNo = bigarray.SearchBatch(0, doesNotExist)
	if dneErr != nil {
		return nil, dneErr
	}
	if r.nextSeqNo > 0 {
		name := packName(prefix, r.nextSeqNo-1)
		o := server.bucket.Object(name)
		_ = o
		// TODO: read object, set lastTimestamp and lastHash
	}

	handler := func(ctx context.Context, v interface{}) {
		entries := v.([]*packRequest)

		// Drop entries with timestamps before r.lastTimestamp
		j := 0
		for _, e := range entries {
			if e.pb.Timestamp.AsTime().Before(r.lastTimestamp) {
				e.ch <- status.Errorf(codes.Aborted, "timestamp too early (req: %s, last: %s)", e.pb.Timestamp.AsTime(), r.lastTimestamp)
			} else {
				entries[j] = e
				j++
			}
		}
		entries = entries[:j]

		sort.Slice(entries, func(i, j int) bool {
			tsi := entries[i].pb.Timestamp.AsTime()
			tsj := entries[j].pb.Timestamp.AsTime()
			return tsi.Before(tsj)
		})

		// TODO: do the packing, calculate pack hash.

		// TODO: submit pack hash for notarization to prefix[:len(prefix)-1]
		// and block until notarized. NB: may have spurious notarizations in
		// higher level prefix tree if writing to bucket below fails. That's ok.
		// The prefix tree attests to a singular sequencing of all log entries
		// virtue of its sequential and tree-shaped hash chaining. The
		// notarization to a higher level prefix tree is only to order a new
		// pack against all other packs in all other prefix trees.

		// TODO: write pack with notarization to bucket with DoesNotExist
		// condition.

		// success
		r.lastTimestamp = entries[len(entries)-1].pb.Timestamp.AsTime()
		for _, r := range entries {
			r.ch <- nil
		}

		// TODO: If top-level (prefix=""), broadcast new PrefixInfo across
		// Serf agent (via ringMux?).
	}
	r.bundler = autobundler.New(ctx, &packRequest{}, handler, maxPackSize)
	return r, nil
}

type packRequest struct {
	pb *pb.PackRequest
	ch chan error
}

type packServer struct {
	ctx    context.Context // for prefixPacker's
	bucket *storage.BucketHandle

	// lots of reads (every RPC handler) and few writes (handling a new prefix)
	packers *sync.Map           // map[string]*prefixPacker
	sf      *singleflight.Group // make packers once (it's slow)

	// For faster channel allocation
	chPool *sync.Pool

	pb.UnimplementedPackerServer
}

func newPackServer(ctx context.Context, bkt *storage.BucketHandle) *packServer {
	r := &packServer{
		ctx:     ctx,
		bucket:  bkt,
		packers: &sync.Map{},
		sf:      &singleflight.Group{},
		chPool:  &sync.Pool{},
	}
	r.chPool.New = func() interface{} {
		return make(chan error)
	}
	return r
}

func (s *packServer) Pack(ctx context.Context, r *pb.PackRequest) (*pb.PackResponse, error) {
	p := prefix.ToString(r.Document, prefix.LengthNibbles)

	// Get the right packer or create it.
	packerI, ok := s.packers.Load(p)
	if !ok {
		var err error
		packerI, err, _ = s.sf.Do(p, func() (interface{}, error) {
			var newPacker *prefixPacker
			newPacker, err := newPrefixPacker(s.ctx, s, p)
			if err != nil {
				return nil, err
			}
			packerI, _ = s.packers.LoadOrStore(p, newPacker)
			return packerI, nil
		})
		if err != nil {
			return nil, err
		}
	}

	packer := packerI.(*prefixPacker)
	req := new(packRequest)
	req.pb = r
	req.ch = s.chPool.Get().(chan error)
	packer.bundler.Add(ctx, req)
	err := <-req.ch
	s.chPool.Put(req.ch)
	if err != nil {
		return nil, err
	}

	// TODO: prepare PackResponse
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
