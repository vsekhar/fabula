package worktoken

import (
	"context"
	"encoding/binary"
	"math"
	"time"

	"golang.org/x/crypto/sha3"
)

const hashOutputLength = 8 // bytes

func Generate(ctx context.Context, b []byte, t time.Time, leadingZeroBits uint, goroutines int) (uint64, error) {
	if goroutines < 1 {
		panic("goroutines must be at least 1")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ch := make(chan uint64)
	shardsize := math.MaxUint64 / uint64(goroutines)
	for i := 0; i < int(goroutines); i++ {
		go func(shard int) {
			q := shardsize * uint64(shard)
			qb := make([]byte, binary.MaxVarintLen64)
			tb := make([]byte, binary.MaxVarintLen64)
			binary.PutVarint(tb, t.UnixNano())
			refh := sha3.NewShake256()
			refh.Write(b)
			refh.Write(tb)
			out := make([]byte, hashOutputLength)
			donech := ctx.Done()
			for ; ; q++ {
				select {
				case <-donech:
					return
				default:
				}
				h := refh.Clone()
				qbn := binary.PutUvarint(qb, q)
				h.Write(qb[:qbn])
				n, _ := h.Read(out)
				remainingZeroBits := leadingZeroBits
				for k := 0; k < n; k++ {
					var mask byte = 255 >> remainingZeroBits
					if out[k]&mask == out[k] {
						if remainingZeroBits <= 8 {
							// found it
							ch <- q
							return
						}
						remainingZeroBits -= 8 // continue checking
					} else {
						break // continue searching
					}
				}
			}
		}(i)
	}
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case v := <-ch:
		return v, nil
	}
}
