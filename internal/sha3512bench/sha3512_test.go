package sha3512_test

import (
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/sha3"
)

func BenchmarkSHA3512(b *testing.B) {
	buf := make([]byte, 64)
	h := sha3.New512()
	for i := 0; i < b.N; i++ {
		rand.Read(buf)
		h.Write(buf)
	}
	h.Sum(buf)
}
