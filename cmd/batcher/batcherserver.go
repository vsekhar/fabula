package main

import (
	"context"
	"sync"

	"github.com/golang/protobuf/descriptor"
	log "github.com/sirupsen/logrus"
	pb "github.com/vsekhar/fabula/internal/api/batcher"
	"github.com/vsekhar/fabula/internal/batcher"
	"github.com/vsekhar/fabula/internal/peerbook"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// keep in sync with internal/api/batcher/batcher.proto
const prefixFieldID = 2

var rootPrefix string

func init() {
	msg := &pb.InternalBatchRequest{}
	_, md := descriptor.MessageDescriptorProto(msg)
	var field *descriptorpb.FieldDescriptorProto
	for _, f := range md.GetField() {
		if *f.Number == prefixFieldID {
			field = f
			break
		}
	}
	if field == nil {
		log.Fatalf("'prefix' field (#%d) not found in InternalBatchRequest proto", prefixFieldID)
	}
	rootPrefix = proto.GetExtension(field.Options, pb.E_Root).(string)
	if rootPrefix == "" {
		log.Fatalf("root prefix is empty, fix batcher.proto")
	}
}

type internalBatcherServer struct {
	peers       *peerbook.PeerBook
	sequencer   *batcher.Batcher
	batcher     *batcher.Batcher
	payloadPool *sync.Pool

	pb.UnimplementedInternalBatcherServer
}

func newInternalBatcherServer(peers *peerbook.PeerBook) *internalBatcherServer {
	var svr *internalBatcherServer
	var nilPayload *batchPayload
	svr = &internalBatcherServer{
		peers:     peers,
		sequencer: batcher.New(nilPayload, svr.sequenceHandler, 1000, 1),
		batcher:   batcher.New(nilPayload, svr.batchHandler, 1000, 50),
		payloadPool: &sync.Pool{New: func() interface{} {
			return &batchPayload{
				ch: make(chan *batchResponse),
			}
		}},
	}
	// TODO: start listener for peerbook broadcasts of high water mark
	return svr
}

func (i *internalBatcherServer) batchHandler(batch interface{}) {
	// TODO: writeToPubSub()
	// TODO: go writeToStorage()
	// TODO: send upward
	// TODO: retry logic (send to servers corresponding to prefix, prefix2, ...)
}

func (i *internalBatcherServer) sequenceHandler(batch interface{}) {
	// TODO: Write to atomic writer
	// TODO: retry logic (send to servers corresponding to *, *2, ...)
}

type batchPayload struct {
	r  *pb.InternalBatchRequest
	ch chan *batchResponse
}

type batchResponse struct {
	resp *pb.InternalBatchResponse
	err  error
}

func (i *internalBatcherServer) InternalBatch(ctx context.Context, r *pb.InternalBatchRequest) (*pb.InternalBatchResponse, error) {
	b := i.batcher
	if r.Prefix == rootPrefix {
		b = i.sequencer
	}

	payload := i.payloadPool.Get().(*batchPayload)
	if err := b.Add(ctx, payload); err != nil {
		// don't return payload to pool, we don't know if the channel has been
		// used.
		return nil, err
	}
	select {
	case resp := <-payload.ch:
		payload.r = nil
		i.payloadPool.Put(payload)
		return resp.resp, resp.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
