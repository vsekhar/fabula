package testlog

// from: https://git.sr.ht/~samwhited/testlog/tree/master/testlog.go

import (
	"log"
	"testing"
)

// New returns a new logger that logs to the provided testing.T.
func New(t testing.TB) *log.Logger {
	t.Helper()
	return log.New(testWriter{TB: t}, t.Name()+" ", log.LstdFlags|log.Lshortfile|log.LUTC)
}

type testWriter struct {
	testing.TB
}

func (tw testWriter) Write(p []byte) (int, error) {
	tw.Helper()
	tw.Logf("%s", p)
	return len(p), nil
}
