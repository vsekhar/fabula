package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/vsekhar/fabula/internal/bigarray"
	"github.com/vsekhar/fabula/internal/prefix"
	"github.com/vsekhar/fabula/pkg/api/storagepb"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/support/bundler"
	"google.golang.org/protobuf/proto"
)

const bundleSize = 1 << (prefix.LengthNibbles * 4)
const delayThreshold = 1 * time.Second

var bundlers sync.Map

func msgSize(msg *pubsub.Message) int {
	s := len(msg.Data) + 20 /* timestamp */ + len(msg.ID) + len(msg.OrderingKey)
	for k, v := range msg.Attributes {
		s += len(k) + len(v)
	}
	return s
}

func newBundler(handler func(interface{})) *bundler.Bundler {
	nb := bundler.NewBundler(&pubsub.Message{}, handler)
	nb.DelayThreshold = delayThreshold
	nb.BundleCountThreshold = bundleSize
	nb.BundleByteThreshold = 0
	nb.BundleByteLimit = 0
	nb.BufferedByteLimit = 1e9 // 1G
	nb.HandlerLimit = 1        // sequential
	return nb
}

func packName(prefix string, seqNo int) string {
	return fmt.Sprintf("%s-%d.pack", prefix, seqNo)
}

func unpackName(name string) (prefix string, seqNo int, err error) {
	cons := func(s, sep string) (h, t string) {
		parts := strings.SplitN(s, sep, 1)
		if len(parts) > 0 {
			h = parts[0]
			if len(parts) > 1 {
				t = parts[1]
			}
		}
		return
	}

	prefix, rest := cons(name, "-")
	seqNos, rest := cons(rest, ".")
	if rest != "pack" {
		return "", 0, fmt.Errorf("missing .pack extension: %s", name)
	}
	s, err := strconv.ParseUint(seqNos, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("cannot parse sequence number in '%s': %s", name, err)
	}
	seqNo = int(s)
	return
}

type handler struct {
	ctx           context.Context
	bucket        *storage.BucketHandle
	orderingKey   string
	nextSeqNo     int
	committedTime time.Time
}

func (h *handler) Handle(i interface{}) {
	msgs := i.([]*pubsub.Message)
	var success bool
	defer func() {
		// Ack in reverse order in case we fail here. Pubsub will redeliver
		// all subsequent messages after a message missing an Ack.
		for i := len(msgs) - 1; i >= 0; i-- {
			if success {
				msgs[i].Ack()
			} else {
				msgs[i].Nack()
			}
		}
	}()

	if len(msgs) > bundleSize {
		log.Fatalf("Bundle size %d, expected max. %d", len(msgs), bundleSize)
	}

	var earliest time.Time = msgs[0].PublishTime
	var last time.Time = msgs[0].PublishTime
	for _, m := range msgs {
		if m.PublishTime.Before(last) {
			log.Fatalf("Messages out of order: %v", msgs)
		}
		if m.OrderingKey != h.orderingKey {
			log.Fatalf("Expected ordering key %s, got %s", h.orderingKey, m.OrderingKey)
		}
		last = m.PublishTime
	}
	if earliest.Before(h.committedTime) {
		log.Fatalf("Bundle contains messages earlier (%s) than already stored (%s)", earliest, h.committedTime)
	}
	log.Printf("Bundle (%d messages): %s to %s", len(msgs), earliest, last)

	doesNotExist := func(i int) bool {
		_, err := h.bucket.Object(packName(h.orderingKey, h.nextSeqNo)).Attrs(h.ctx)
		if err == storage.ErrObjectNotExist {
			return true
		}
		if err == nil {
			return false
		}
		log.Fatal(err)
		return false
	}
	if h.nextSeqNo == 0 {
		// likely a new handler, first write is likely to fail, so probe
		h.nextSeqNo = bigarray.Search(h.nextSeqNo, doesNotExist)
	}
	tries := 3
	for i := 0; i < tries; i++ {
		prevPackName := packName(h.orderingKey, h.nextSeqNo-1)
		r, err := h.bucket.Object(prevPackName).NewReader(h.ctx)
		if err == storage.ErrObjectNotExist {
			log.Fatalf("expected... ")
		}
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		r.Close()
		if err != nil {
			log.Fatal(err)
		}
		m := &storagepb.PackStorage{}
		err = proto.Unmarshal(b, m)
		if err != nil {
			log.Fatal(err)
		}

		// Verify previous pack entries are in order
		lastOfPrev := m.Entries[0].GetTimestamp().AsTime()
		for _, e := range m.Entries {
			et := e.GetTimestamp().AsTime()
			if et.Before(lastOfPrev) {
				log.Fatalf("entries out of order in %s: %v", prevPackName, m.Entries)
			}
			lastOfPrev = et
		}

		// Verify previous pack does not contain any of the current entries.
		// This can occur if we failed after writing the previous pack but
		// before Acking all the entries to PubSub.
		//
		// If the previous pack contains a prefix of our entries with no gaps,
		// we can Ack those entries, drop them from the current batch and
		// persist the rest.

		// hash previous and children to compute mmr_SHA3512
		// TODO: pack cache

		// Assemble new pack
		n := &storagepb.PackStorage{}
		b, err = proto.Marshal(n)
		if err != nil {
			log.Fatal(err)
		}

		// Notarize new pack with a higher-level (shorter) orderingKey
		//
		// NB: it's ok if we fail after notarization. The pack storage (and the
		// sequence numbers in their names) represents the canonical and
		// transactional record of its prefix chain. If this pack later fails
		// to be written, another (notarized) pack will take its place. The
		// notarization of the failed pack will be orphaned but is otherwise harmless.

		o := h.bucket.Object(packName(h.orderingKey, h.nextSeqNo)).If(storage.Conditions{DoesNotExist: true})
		w := o.NewWriter(h.ctx)
		_, err = w.Write(b)
		if err != nil {
			log.Fatal(err)
		}
		if err := w.Close(); err != nil {
			// If we failed to write because it already exists, update
			switch e := err.(type) {
			case *googleapi.Error:
				if e.Code == http.StatusPreconditionFailed {
					h.nextSeqNo = bigarray.Search(h.nextSeqNo, doesNotExist)
					continue
				}
			default:
				log.Fatal(err)
			}
		}
		success = true
		break
	}
}

func pack(ctx context.Context) {
	defer func() {
		// flush bundlers
		bundlers.Range(func(key, value interface{}) bool {
			b := value.(*bundler.Bundler)
			b.Flush()
			return true
		})
	}()

	sub := getSubscriptionOrCreateOrDie(ctx, "all")
	bkt := getBucketOrDie(ctx, *bucket)

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		bi, present := bundlers.Load(msg.OrderingKey)
		if !present {
			nh := new(handler)
			nh.ctx = ctx
			nh.bucket = bkt
			nh.orderingKey = prefix.ToString(msg.Data, prefix.LengthNibbles)
			// Use LoadOrStore to prevent a race condition between Load above
			// and a Store here.
			nb := newBundler(func(i interface{}) { nh.Handle(i) })
			bi, _ = bundlers.LoadOrStore(msg.OrderingKey, nb)
		}
		// TODO: try skipping msgSize, we don't set a size threshold in
		// the bundlers.
		err := bi.(*bundler.Bundler).AddWait(ctx, msg, msgSize(msg))
		if err != nil {
			msg.Nack()
			log.Print(err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
