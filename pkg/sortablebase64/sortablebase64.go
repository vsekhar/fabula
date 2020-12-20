// Package sortablebase64 contains routines for encoding text in a variant of
// base64. This variant has the useful property that encodings of numeric
// values retain their ordering under lexicographic (ASCII) sort as the values
// do under numeric sort.
package sortablebase64

import (
	"fmt"
	"strings"
)

// Inspiration: https://www.codeproject.com/Articles/5165340/Sortable-Base64-Encoding

// Alphabet is an ordered list of characters used in sortablebase64 encoding.
// The order is drawn from the placement of these characters in the ASCII table.
const Alphabet = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

var decodeMap [256]byte

func init() {
	alphamap := make(map[rune]struct{})
	for i, c := range Alphabet {
		if i > 0 {
			if Alphabet[i-1] >= Alphabet[i] {
				panic(fmt.Sprintf("bad alphabet order: %c < %c", Alphabet[i], Alphabet[i-1]))
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
	for i := 0; i < 11; i++ {
		rb.WriteByte(Alphabet[n>>(60-(i*6))&0x3F])
	}
	return rb.String()
}

// DecodeUint64 returns a uint64 representing the sortablebase64-encoded string
// s, or an error.
func DecodeUint64(s string) (uint64, error) {
	if len(s) != 11 {
		return 0, fmt.Errorf("sortablebase64: expected string of length 11, got string of length %d", len(s))
	}
	b := []byte(s)
	var u uint64
	for i := 0; i < 11; i++ {
		v := uint64(decodeMap[b[i]])
		if v == 0xFF {
			return 0, fmt.Errorf("illegal character at pos %d in '%s'", i, s)
		}
		u |= v << (60 - (i * 6))
	}
	return u, nil
}

// IncUint64 increments the sortablebase64-encoded uint64 s.
func IncUint64(s string) (string, error) {
	u, err := DecodeUint64(s)
	if err != nil {
		return "", err
	}
	return EncodeUint64(u + 1), nil
}

// DecUint64 increments the sortablebase64-encoded uint64 s.
func DecUint64(s string) (string, error) {
	u, err := DecodeUint64(s)
	if err != nil {
		return "", err
	}
	return EncodeUint64(u - 1), nil
}
