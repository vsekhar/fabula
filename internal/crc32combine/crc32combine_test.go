package crc32combine_test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/vsekhar/fabula/internal/crc32combine"
)

func TestCombineRand(t *testing.T) {
	const poly = crc32.IEEE // also: crc32.Castagnoli

	// buf := []byte{0, 148, 190, 58, 99, 42, 255, 0, 0, 0} // even length
	buf := make([]byte, 1024)
	rand.Read(buf)
	b1, b2 := buf[:len(buf)/2], buf[len(buf)/2:]

	table := crc32.MakeTable(poly)

	h := crc32.New(table)
	h.Write(b1)
	c1 := binary.BigEndian.Uint32(h.Sum(nil))

	h = crc32.New(table)
	h.Write(b2)
	c2 := binary.BigEndian.Uint32(h.Sum(nil))

	h = crc32.New(table)
	h.Write(buf)
	ec := binary.BigEndian.Uint32(h.Sum(nil))

	c := crc32combine.Combine(c1, c2, len(b2), poly)
	if ec != c {
		t.Errorf("got %d, expected %d", c, ec)
	}
}

func TestGCSCompose(t *testing.T) {
	const poly = crc32.Castagnoli // per GCS

	// https://github.com/googleapis/googleapis/blob/6ae2d424245deeb34cf73c4f7aba31f1079bcc40/google/api/annotations.proto
	// gsutil hash annotations.proto
	a, _ := base64.StdEncoding.DecodeString("Oa2A2A==")

	// https://github.com/googleapis/googleapis/blob/ca1372c6d7bcb199638ebfdb40d2b2660bab7b88/google/api/http.proto
	// gsutil hash http.proto
	b, _ := base64.StdEncoding.DecodeString("A1kEDg==") // len = 15140

	// cat annotations.proto http.proto > combined.proto
	// gsutil hash combined.proto
	c, _ := base64.StdEncoding.DecodeString("rKtODg==")

	ai := binary.BigEndian.Uint32(a)
	bi := binary.BigEndian.Uint32(b)
	ci := binary.BigEndian.Uint32(c)
	ec := crc32combine.Combine(ai, bi, 15140, poly)
	if ec != ci {
		t.Errorf("got %d, expected %d", ci, ec)
	}
}
