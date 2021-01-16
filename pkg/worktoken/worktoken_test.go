package worktoken_test

import (
	"context"
	"encoding/binary"
	"runtime"
	"testing"
	"time"

	"github.com/vsekhar/fabula/pkg/worktoken"
)

func TestGenerate(t *testing.T) {
	const leadingZeroBits = 16
	const rounds = 8
	goroutines := runtime.NumCPU() * 2

	ctx := context.Background()
	d := []byte("abc")
	for i := 0; i < rounds; i++ {
		c, err := worktoken.Generate(ctx, d, time.Now(), leadingZeroBits, goroutines)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("c[%d]: %d", i, c)
		cb := make([]byte, binary.MaxVarintLen64)
		cbn := binary.PutUvarint(cb, c)
		d = append(d, cb[:cbn]...)
	}
	t.Error("output")
}
