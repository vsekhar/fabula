// Package mmr emulates a Merkle Mountain Range service.
package mmr

import (
	"crypto/rand"
	"time"

	"github.com/vsekhar/COMMIT/faketime"
	"golang.org/x/crypto/sha3"
)

const (
	SALTSIZE = 64
)

type MMR struct {
	a [][]byte
}

func New() *MMR {
	return &MMR{
		a: make([][]byte, 0),
	}
}

func (m *MMR) Add(b []byte) (salt []byte, t time.Time, i int) {
	t, _ = faketime.Nowish()
	i = len(m.a)
	salt = make([]byte, SALTSIZE)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	h := sha3.NewShake256()
	// hash children
	h.Write(b)
	h.Write(salt)
	faketime.SleepUntilPast(t)
}
