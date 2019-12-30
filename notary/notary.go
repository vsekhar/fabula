// Package notary is a proof-of-concept public notary service.
package notary

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"time"
)

const saltLength = 64

// Factored out to ensure we assemble payloads the same way when signing and verifying
func assembleLeaf(b, salt []byte, t time.Time) []byte {
	buf := new(bytes.Buffer)
	buf.Write(b)
	buf.Write(salt)
	binary.Write(buf, binary.LittleEndian, t)
	return buf.Bytes()
}

// Notarization contains the data returned by a notary service.
type Notarization struct {
	Salt      []byte
	Timestamp time.Time
	Signature []byte
	PublicKey ed25519.PublicKey
}

// ProofEntry contains slices of bytes that should be pre- and post-pended to a given
// data element before generating its signature.
type ProofEntry struct {
	pre  [][]byte
	post [][]byte
}

// Proof is a slice of entries that represent each of the hashes used to generate
// the head
type Proof []ProofEntry

// Log is a full log of all notarizations, including tree elements used in generating
// compact proofs.
type Log [][]byte

// Digest is a hash summarizing the notary log.
type Digest []byte
