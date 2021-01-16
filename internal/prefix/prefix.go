package prefix

import (
	"encoding/hex"
	"log"
)

// Number of nibbles: number of unique prefixes
//
//     8: 4.3B
//     7: 268M
//     6: 16M
//     5: 1M
//     4: 65K
//     3: 4K
//     2: 256
//     1: 16
//     0: 1

// LengthNibbles is the number of nibbles (4 bytes) considered the prefix
// of a given hash.
const LengthNibbles = 5 // 1M prefixes

// ToString returns the prefix of b, of ceil(nibbles/2)length, as a canonical
// string.
func ToString(b []byte, nibbles int) string {
	orderingPrefixBytes := (LengthNibbles / 2) + 1
	return hex.EncodeToString(b[:orderingPrefixBytes])[:nibbles]
}

// FromString returns the binary prefix
func FromString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		log.Fatalf("error decoding prefix string %s", s)
	}
	return b
}
