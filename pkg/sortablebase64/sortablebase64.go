// Package sortablebase64 contains routines for encoding text in a variant of
// base64. This variant has the useful property that encodings of numeric
// values retain their ordering under lexicographic (ASCII) sort as the values
// do under numeric sort.
package sortablebase64

import (
	"log"
	"strings"
)

// Inspiration: https://www.codeproject.com/Articles/5165340/Sortable-Base64-Encoding

const Alphabet = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

var decodeMap [256]byte

func init() {
	alphamap := make(map[rune]struct{})
	for i, c := range Alphabet {
		if i > 0 {
			if Alphabet[i-1] >= Alphabet[i] {
				log.Fatalf("bad alphabet order: %c < %c", Alphabet[i], Alphabet[i-1])
			}
		}
		alphamap[c] = struct{}{}
	}
	if len(alphamap) != 64 {
		panic("bad alphabet, must be 64 non-duplicated chars")
	}

	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = 0xFF
	}
	for i := 0; i < len(Alphabet); i++ {
		decodeMap[Alphabet[i]] = byte(i)
	}
}

// EncodeUint64 returns a string representing n in sortablebase64.
//
// The returned string is always 11 bytes in length.
func EncodeUint64(n uint64) string {
	rb := strings.Builder{}

	// Write big-endian
	for i := 0; i < 11; i++ {
		rb.WriteByte(Alphabet[(n>>(60-(i*6)))&0x3F])
	}
	/*
		rb.WriteByte(Alphabet[n>>56&0x3F])
		rb.WriteByte(Alphabet[n>>48&0x3F])
		rb.WriteByte(Alphabet[n>>42&0x3F])
		rb.WriteByte(Alphabet[n>>36&0x3F])
		rb.WriteByte(Alphabet[n>>30&0x3F])
		rb.WriteByte(Alphabet[n>>24&0x3F])
		rb.WriteByte(Alphabet[n>>18&0x3F])
		rb.WriteByte(Alphabet[n>>12&0x3F])
		rb.WriteByte(Alphabet[n>>6&0x3F])
		rb.WriteByte(Alphabet[n&0x3F]) // LSB
	*/
	return rb.String()
}

// DecodeUint64 returns a uint64 representing the sortablebase64-encoded string
// s, or an error.
func DecodeUint64(s string) uint64 {
	var u uint64
	if len(s) != 11 {
		log.Fatalf("expected string of length 11, got string of length %d", len(s))
	}
	b := []byte(s)
	for i := 0; i < 11; i++ {
		v := uint64(decodeMap[b[i]])
		if v == 0xFF {
			log.Fatalf("illegal character at pos %d in '%s'", i, s)
		}
		shift := 60 - (6 * i)
		shiftedV := v << shift
		u = u | shiftedV
	}
	return u
}

// Inc increments a sortablebase64 in string form without decoding the entire
// value.
func Inc(s string) string {
	if len(s) != 11 {
		log.Fatalf("expected string of length 11, got string of length %d", len(s))
	}
	b := []byte(s)
	rb := make([]byte, len(b))
	inc := true
	for i := 10; i >= 0; i-- {
		v := int(decodeMap[b[i]])
		if v == 0xFF {
			log.Fatalf("illegal character at pos %d in '%s'", i, s)
		}
		if inc {
			v = (v + 1) % len(Alphabet)
			if v != 0 {
				inc = false
			}
		}
		rb[i] = Alphabet[v]
	}
	if inc == true {
		panic("overflow")
	}
	return string(rb)
}

// Dec decrements a sortablebase64 in string form without decoding the entire
// value.
func Dec(s string) string {
	if len(s) != 11 {
		log.Fatalf("expected string of length 11, got string of length %d", len(s))
	}
	b := []byte(s)
	rb := make([]byte, len(b))
	dec := true
	for i := 10; i >= 0; i-- {
		v := int(decodeMap[b[i]])
		if v == 0xFF {
			log.Fatalf("illegal character at pos %d in '%s'", i, s)
		}
		if dec {
			if v == 0 {
				v = len(Alphabet) - 1
			} else {
				v = v - 1
				dec = false
			}
		}
		rb[i] = Alphabet[v]
	}
	if dec == true {
		panic("underflow")
	}
	return string(rb)
}
