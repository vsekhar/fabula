package notary

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"io"

	"github.com/vsekhar/COMMIT/internal/truetimeish"
)

// Service is a verifiable notary service.
type Service struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	log        [][]byte // MMR
}

// NewService returns a new notary Service.
func NewService() (*Service, error) {
	return newServiceFromRand(nil)
}

func newServiceFromRand(rand io.Reader) (*Service, error) {
	n := new(Service)
	var err error
	n.publicKey, n.privateKey, err = ed25519.GenerateKey(rand)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// Key returns the public key of the Service.
func (s *Service) Key() []byte {
	return s.publicKey
}

// Notarize notarizes data and returns the timestamp, salt and signature, or an error.
func (s *Service) Notarize(b []byte) (n Notarization, err error) {
	return s.notarizeImpl(ProofEntry{}, b)
}

func (s *Service) notarizeImpl(p ProofEntry, b []byte) (n Notarization, err error) {
	ts := truetimeish.Get()
	// TODO: call self for non-leaf nodes
	n.Salt = make([]byte, saltLength)
	rand.Read(n.Salt)
	n.Signature, err = s.privateKey.Sign(nil, assembleLeaf(b, n.Salt, ts.Timestamp()), crypto.Hash(0))
	if err != nil {
		return Notarization{}, err
	}
	n.Timestamp = ts.Timestamp()
	n.PublicKey = s.publicKey
	s.log = append(s.log, n.Signature)
	return n, nil
}

// Log returns the full notary log and its digest.
func (s *Service) Log() Log {
	return s.log
}

// Digest returns a hash summarizing the service log.
func (s *Service) Digest() Digest {
	// bag peaks
	panic("unimplemented")
}

// Prove returns a Proof that can be used to verify that sig has been incorporated
// into the service log, or nil if no such proof can be generated.
func (s *Service) Prove(sig []byte) *Proof {
	panic("unimplemented")
}

// ProveDigest returns a Proof that can be used to verify that digest a is a
// predecessor and covered by digest b, or nil if no such proof can be generated.
func (s *Service) ProveDigest(a, b []byte) *Proof {
	panic("unimplemented")
}
