// Package faketime emulates a TrueTime service.
package faketime

import (
	"math"
	"math/rand"
	"time"
)

// Epsilon is the time uncertainty used by functions in package faketime.
//
// Changing Epsilon while concurently calling functions in package faketime may yield
// undefined behavior.
var Epsilon = 10 * time.Millisecond

// epislon returns a random positive duration drawn from a rectified normal
// distribution with mean 0 and standard deviation jitterMs.
func epsilon(jitterMs float64) time.Duration {
	jitterNano := jitterMs * 1000
	return time.Duration(math.Abs(rand.NormFloat64() * jitterNano))
}

// Nowish returns a time interval within which the current time resides.
//
// The interval is best understood by its exclusion property: the current time
// is guaranteed not to be earlier than earliest and the current time is
// guaranteed not to be later than latest. Nowish makes no other claims about the
// current time.
func Nowish() (earliest, latest time.Time) {
	t := time.Now()
	return t.Add(-Epsilon), t.Add(Epsilon)
}

// CausalNow returns a recent time that is guaranteed to be prior to the timestamp
// subsequently returned by any causally-connected future call to Nowish or CausalNow.
//
// CausalNow performs a commit-wait before returning. Use Nowish and SleepUntilPast if
// you'd like to perform work concurrently with the commit-wait.
func CausalNow() time.Time {
	t := time.Now()
	time.Sleep(Epsilon)
	return t
}

// SleepUntilPast sleeps until t is for sure in the past.
func SleepUntilPast(t time.Time) {
	e, _ := Nowish()
	for e.Sub(t) > 0 {
		time.Sleep(e.Sub(t))
		e, _ = Nowish()
	}
}
