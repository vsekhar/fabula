package maglevbench

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/dgryski/go-maglev"
	"github.com/golang/groupcache/consistenthash"
)

var serverCounts = []int{
	1, 10, 100, 1000, 10000,
}

const consistenthashTargetEntries = 10000

var (
	names      []string
	lookupKeys []string
)

func init() {
	names = make([]string, 100000000)
	for i := range names {
		names[i] = fmt.Sprintf("backend-%d", i)
	}
	lookupKeys = make([]string, 1<<10) // large enough to avoid cache hits
	for i := range lookupKeys {
		lookupKeys[i] = fmt.Sprintf("key-%d", i)
	}
}

func BenchmarkConsistenHashByServerCount(b *testing.B) {
	for _, s := range serverCounts {
		b.Run(fmt.Sprintf("%d_servers", s), func(b *testing.B) {
			replicas := consistenthashTargetEntries / s
			if replicas < 1 {
				replicas = 1
			}
			for i := 0; i < b.N; i++ {
				ch := consistenthash.New(replicas, nil)
				for _, n := range names[:s] {
					ch.Add(n)
				}
			}
		})
	}
}

func BenchmarkConsistentHashLookup(b *testing.B) {
	for _, s := range serverCounts {
		b.Run(fmt.Sprintf("%d_servers", s), func(b *testing.B) {
			replicas := consistenthashTargetEntries / s
			if replicas < 1 {
				replicas = 1
			}
			ch := consistenthash.New(replicas, nil)
			for _, n := range names[:s] {
				ch.Add(n)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ch.Get(lookupKeys[rand.Intn(len(lookupKeys))])
			}
		})
	}
}

func BenchmarkMaglevByServerCount(b *testing.B) {
	f := func(b *testing.B, m uint64) {
		for _, s := range serverCounts {
			b.Run(fmt.Sprintf("%d_servers", s), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = maglev.New(names[:s], m)
				}
			})
		}
	}

	b.Run("BigM", func(b *testing.B) { f(b, maglev.BigM) })
	b.Run("SmallM", func(b *testing.B) { f(b, maglev.SmallM) })
}

func BenchmarkMaglevLookup(b *testing.B) {
	f := func(b *testing.B, m uint64) {
		for _, s := range serverCounts {
			b.Run(fmt.Sprintf("%d_servers", s), func(b *testing.B) {
				m := maglev.New(names[:s], m)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = m.Lookup(rand.Uint64())
				}
			})
		}
	}

	b.Run("BigM", func(b *testing.B) { f(b, maglev.BigM) })
	b.Run("SmallM", func(b *testing.B) { f(b, maglev.SmallM) })
}
