// Package streamreduce provides streaming MapReduce semantics on top of
// Google Cloud Pubsub with the message ordering feature.
package streamreduce

// TODO: nix all the pubsub stuff, make it abstract.
//  - user adds messages
//  - user provides a function that takes a list of messages and produces a
//    list of messages (this function should not call Add()).

// TODO: add auto rate bundling depending on execution time
//  - estimate three values
//  - in rate = estimated rate at which messages are coming in
//  - out rate = fixed + variable(n) rate at which f() processes messages
//
// Find n such at: in rate == out rate == g(fixed, variable(n))
//
// Start by processing first message right away, get {1, t_fixed+t_var}
// IF a message arrived while processing first message, then try bundle of 2.
//
// Keep submitting values {n, t_fixed+n*t_var} to build model to get t's.
//
// Then use t's to find ideal n =
//
// Don't buffer, block when messages get backed up.
import (
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/support/bundler"
)

// Default parameters
const (
	DefaultDelayThreshold       = 500 * time.Millisecond
	DefaultBundleCountThreshold = 50
)

const (
	none = iota
	nack
	ack
)

// Message is a message to be reduced.
type Message struct {
	// This struct is to proxy acks and nacks so that we can reply based on
	// whether new messages were sent successfully after bumping the ordering
	// key.

	m      *pubsub.Message
	result int

	Data        []byte
	Attributes  map[string]string
	OrderingKey string

	ID              string
	PublishTime     time.Time
	DeliveryAttempt *int
}

// Ack indicates successful processing of a message.
func (m *Message) Ack() {
	if m.result == none {
		m.result = ack
	}
}

// Nack indicates unsuccessful processing of a message.
func (m *Message) Nack() {
	if m.result == none {
		m.result = nack
	}
}

// StreamReducer is the interface a client of streamreducer must implement to
// produce KeyReducers for a given ordering_key.
//
// KeyReducers will be cached and reused for all messages with a given
// ordering_key.
type StreamReducer interface {
	// NewReducerFor returns a KeyReducer for ordering_key key. NewReducerFor
	// will only be called once for any given key.
	NewReducerFor(key string) KeyReducer

	// TopKey returns the top-level key in the stream's hierarchical keyspace.
	// The top level key cannot be the empty string.
	//
	// The client must ensure that the key space is truly hierarchical. That is,
	// calling NextKey() on a given KeyReducer, creating a KeyReducer for that
	// key and calling NextKey() again will eventually produce the TopKey. If
	// not, messages may enter a continuous loop.
	TopKey() string
}

// KeyReducer is the interface a client of streamreducer must implement to
// reduce a specific key.
type KeyReducer interface {
	// Reduce reduces in to zero or messages out. Reduce should call Ack or Nack
	// on each input message as appropriate.
	//
	// Reduce will only ever be called for messages of a fixed key. Messages may
	// appear more than once to Reduce. If this occurs, messages presented to
	// Reduce will always be presented in the same order though not necessarily
	// with the same batching. As a result, Reduce should be idempotent on a
	// per message (not just a per batch) basis.
	Reduce(in []*Message) (out []*Message)

	// NextKey returns the next higher up key relative to the key of this
	// KeyReducer. NextKey will not be called if this reducer is of the top
	// level key.
	NextKey() string
}

type krImpl struct {
	key     string
	userKR  KeyReducer
	bundler *bundler.Bundler
}

// Instance implements distributed stream reduction via a Cloud PubSub
// topic and subscription. The topic and subscription should be created before
// using Instance.
type Instance struct {

	// Input subscription. Message ordering must be enabled on the subscription
	// and messages must have ordering_key set in a way that evenly distributes
	// traffic.
	Input *pubsub.Subscription

	// A topic to which interim results will be published. This can be the same
	// topic as used for the Input subscription.
	//
	// For interim results reduced from given KeyReducer kr:
	//  ordering_key = kr.NextKey()
	//  attributes.streamreduce_interim_key = kr.NextKey()
	//
	// The topic must have message ordering enabled.
	AggTopic *pubsub.Topic

	// A subscription from which interim results will be read. This subscription
	// must be attached to AggTopic and must have message ordering enabled.
	// AggSubscription can be the same as input.
	AggSubscription *pubsub.Subscription

	// A topic to which output results will be written. This can be the same as
	// AggTopic.
	//
	// Results will be published to Output with:
	//  ordering_key = top_key
	//  (attributes.streamreduce_interim_key will be absent)
	Output *pubsub.Topic

	DelayThreshold       time.Duration
	BundleCountThreshold int

	// TODO use sync.map to make this concurrency-safe, sync.once to create
	//
	keyreducers map[string]KeyReducer
}

// Put puts a new message into the Instance
func (sr *Instance) Put(data []byte, attributes map[string]string, key string) error {
	panic("unimplemented")
}

// TODO: place key in ordering_key and attributes.streamreduce_key. The second
// is so that we can filter on the key (e.g. the top level empty key only) and
// read the aggregated output. Subscriptions can't filter on ordering_key.

// TODO: call streamreducer.Key(k) as soon as messages come in for a given
// ordering_key. That way the work to create a KeyReducer can run in paralle
// with waiting for a batch of messages to fill.

// TODO: can I batch messages with an ordering_key? Will sub.Receive(f(msg))
// give me the next message after f(msg) returns (good), or will it wait until
// msg.Ack() (not good). If the latter, then I need to use the low-level
// streaming API at cloud.google.com/go/pubsub/apiv1.
//  - Actually yes pubsub will provide a batch of messages via sub.Receive()
//    See: https://cloud.google.com/pubsub/docs/pull#concurrency_control
//  - Presumably only one concurrent call to f(msg) will be made at a time to
//    preserve ordering, so submit those messages to a batching queue with the
//    right ordering (e.g. via a channel).
//    - Yes: https://cloud.google.com/pubsub/docs/pull#synchronous_pull
//    - "Pub/Sub delivers a list of messages. If the list has multiple messages,
//      Pub/Sub orders the messages with the same ordering key."
//    - Also: https://cloud.google.com/pubsub/docs/pull#streamingpull
//    - "You provide a callback to the subscriber and the subscriber
//      asynchronously runs the callback for each message. If a subscriber
//      receives messages with the same ordering key, the client libraries
//      sequentially run the callback."
