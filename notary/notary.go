// Package notary is a proof-of-concept public notary service.
package notary

import (
	"crypto/ed25519"
	"io"
	"time"
)

// Service is a verifiable notary service.
type Service struct {
	PublicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// NewService returns a new notary Service.
func NewService() (*Service, error) {
	return newServiceFromRand(nil)
}

func newServiceFromRand(rand io.Reader) (*Service, error) {
	n := new(Service)
	var err error
	n.PublicKey, n.privateKey, err = ed25519.GenerateKey(rand)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// Notarize notarizes bytes
func (n *Service) Notarize(b []byte) (t time.Time, signature, salt []byte) {
	return time.Now(), nil, nil
}

// Client represents a client of a verifiable notary service.
type Client struct {
	svc *Service
}

// NewClient returns a new notary Client.
func NewClient(svc *Service) *Client {
	return &Client{
		svc: svc,
	}
}

// TODO: fetch log, verify it
