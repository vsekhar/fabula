package autobundler

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// addValues adds r random integer values per second to autobundler a until the
// context is cancelled.
func addValues(ctx context.Context, r int, a *AutoBundler) {
	interval := time.Duration(float64(time.Second) / float64(r))
	for i := 0; true; i++ {
		a.Add(ctx, i)
		sleep(ctx, interval)
	}
}

type testCase struct {
	rate     int
	fixed    time.Duration
	variable time.Duration
	max      int
}

func expectedBundle(rate int, fixed, variable time.Duration) int {
	f := float64(fixed) / float64(time.Second)
	v := float64(variable) / float64(time.Second)
	return int(math.Round((float64(rate) * f) / (1.0 - (float64(rate) * v))))
}

func within(a, b int, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	if float64(diff)/float64(b) < tolerance {
		return true
	}
	return false
}

var cases []testCase = []testCase{
	// input rate(/s), fixed time, variable time, max buffer
	{4, 1 * time.Millisecond, 2 * time.Millisecond, 100},

	// Cloud storage has ~300ms roundtrip time to write an object.
	// Sha3512 benchmarks at 1244ns/op.
	// Transferring 64 bytes @ 100MB/s takes 640ns/op.
	// Some sundry memory copying at 1XXns/op.
	// = 300ms + 2us/op
	{1, 300 * time.Millisecond, 2 * time.Microsecond, 100},
	{10, 300 * time.Millisecond, 2 * time.Microsecond, 100},
	{100, 300 * time.Millisecond, 2 * time.Microsecond, 100},
	{100, 300 * time.Millisecond, 2 * time.Microsecond, 1000},
	{500, 300 * time.Millisecond, 2 * time.Microsecond, 1000},
	{10000, 300 * time.Millisecond, 2 * time.Microsecond, 10},

	// cannot reach steady state with 5s settleTime
	// {100, 100 * time.Millisecond, 200 * time.Millisecond, 100},
	// {100, 100 * time.Millisecond, 200 * time.Millisecond, 1000},
	// {10000, 300 * time.Millisecond, 2 * time.Microsecond, 10000},
}

func TestAutoBundler(t *testing.T) {
	const settleTime = 5 * time.Second

	for i, tc := range cases {
		tc, caseNo := tc, i // capture vars
		name := fmt.Sprintf("fix:%s,var:%s,max:%d", tc.fixed, tc.variable, tc.max)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())

			largestBundle := 0
			handler := func(ctx context.Context, v interface{}) {
				b := v.([]int)
				n := len(b)
				if n > largestBundle {
					largestBundle = n
				}
				for i, v := range b {
					if i == 0 {
						continue
					}
					if v != b[i-1]+1 {
						t.Errorf("Case %d: bad sequence @ bundle %d: %d, %d", caseNo, i-1, b[i-1], v)
					}
				}
				sleep(ctx, tc.fixed+(time.Duration(n)*tc.variable))
			}

			a := New(ctx, int(0), handler, tc.max)
			go addValues(ctx, tc.rate, a)
			time.Sleep(settleTime)
			cancel()
			a.Wait()
			eb := expectedBundle(tc.rate, tc.fixed, tc.variable)
			if eb < 0 {
				// pathalogical, will grow to max and then apply back pressure
				eb = tc.max
			}
			if eb < 1 {
				eb = 1
			}
			if eb > tc.max {
				eb = tc.max
			}
			tolerance := 0.1
			if !within(largestBundle, eb, tolerance) {
				t.Errorf("Case %d: largest bundle: %d, expected %d within %.2f%%", caseNo, largestBundle, eb, tolerance*100)
			}
		})
	}
}
