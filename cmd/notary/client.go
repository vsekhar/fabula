package notary

import (
	"crypto/ed25519"
)

// ValidateNotarization returns true if sig is a valid signature generated by key
// using b, t, and salt.
func ValidateNotarization(b []byte, n Notarization) bool {
	msg := assembleLeaf(b, n.Salt, n.Timestamp)
	return ed25519.Verify(n.PublicKey, msg, n.Signature)
}

// ValidateProof returns true if p is a valid proof that sig has been incorporated
// into a notary log.
func ValidateProof(sig []byte, p Proof) {
	for _, e := range p {
		for _, x := range e.pre {
			_ = x
		}
		// hash entity
		for _, x := range e.post {
			_ = x
		}
	}
}

// ValidateLog validates a notary log and its digest.
func ValidateLog(l Log, d Digest) bool {
	panic("unimplemented")
}