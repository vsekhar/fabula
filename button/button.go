// Package button is a proof of concept button service for maintaining
// liveness in an atomic consensus protocol.
package button

// TODO: interaction with notary service. Do clients notarize their pushes? Does the
// button? Button stores audit trail? Record of notarized pushes?

// TODO: timeout. Lazy discovery of timeout condition upon request (using audit trail)?

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/vsekhar/COMMIT/notary"
)

// Button is a one-way atomic primitive.
type Button struct {
	id      []byte
	notary  *notary.Service
	timeout time.Time
	keys    map[string]bool // hex encoded
}

// New returns a new Button backed by notary service n with timeout d and requiring
// "pushes" from the holders of the provided public keys.
//
// A button created with no initial keys can never have its state changed.
//
// Obtaining the state of a button requires Push-ing it.
func New(n *notary.Service, d time.Duration, keys []ed25519.PublicKey) (*Button, error) {
	// log button creation, get timestamp from notary
	nt, err := n.Notarize([]byte{})
	if err != nil {
		return nil, err
	}
	b := &Button{
		// id: signature of notarization,
		notary:  n,
		timeout: nt.Timestamp.Add(d),
		keys:    make(map[string]bool),
	}
	for _, k := range keys {
		b.keys[hex.EncodeToString(k)] = false
	}
	return b, nil
}

func (b *Button) status() bool {
	for _, v := range b.keys {
		if !v {
			return false
		}
	}
	return true
}

// Add adds key to b. If b is not already true, b must be pushed by the holder
// of key in order to become true.
func (b *Button) Add(key ed25519.PublicKey) {
	b.keys[hex.EncodeToString(key)] = false
}

// Push "pushes" the button by the holder of key.
//
// Additional public keys can be specified in addlKeys. These keys will be added
// to the button if they are not already included.
//
// If the button is waiting on no further keys (after pushing with key and adding
// addlKeys), then the button's status is changed from false to true.
//
// Push returns the button's status. If an error occurs, Push returns false and an error.
func (b *Button) Push(msg, sig []byte, key ed25519.PublicKey) (bool, error) {
	if !ed25519.Verify(key, msg, sig) {
		return false, fmt.Errorf("msg/sig/key did not verify correctly")
	}
	strKey := hex.EncodeToString(key)
	if _, ok := b.keys[strKey]; !ok {
		return false, fmt.Errorf("not authorized")
	}
	b.keys[strKey] = true
	return b.status(), nil
}

// AddAndPush adds addlKeys to the button, then pushes using key.
func (b *Button) AddAndPush(msg, sig []byte, key ed25519.PublicKey, addlKeys []ed25519.PublicKey) (bool, error) {
	for _, k := range addlKeys {
		b.Add(k)
	}
	return b.Push(msg, sig, key)
}
