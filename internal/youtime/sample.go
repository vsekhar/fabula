package youtime

import (
	"container/list"
	"math"
	"time"
)

const sigmas = 3 // 99.7% of values

// a sample is a single instance connecting the local relative clock with a
// real wall clock.
type sample struct {
	then time.Time
	rel  relMoment
}

// skewPPB compares two samples from the same wall clock. It returns the number
// of extra the wall clock makes for each billion ticks of relNow.
//
// Calling skewPPB on samples drawn from different wall clocks is undefined.
func (s sample) skewPPB(x sample) int64 {
	interval := s.rel.to(x.rel)
	expected := s.then.Add(interval)
	extra := x.then.Sub(expected)
	extraPPB := extra.Nanoseconds() * (1e9 / interval.Nanoseconds())
	return extraPPB
}

// offset returns a naive difference time offset between two wall clocks sampled
// with s and x. The returned value is naive because it does not account for
// clock skew. I.e. it assumes the two wall clock and the local relative clock
// all tick at exactly the same frequency.
func (s sample) offset(x sample) time.Duration {
	interval := s.rel.to(x.rel)
	expected := s.then.Add(interval)
	return x.then.Sub(expected)
}

// sampleList wraps container/list to handle casting.
type sampleList struct {
	list *list.List
}

func (s *sampleList) len() int {
	return s.list.Len()
}

func (s *sampleList) addAndShift(new sample, max int) {
	s.list.PushBack(new)
	for s.list.Len() > max {
		s.list.Remove(s.list.Front())
	}
}

func (s *sampleList) forEach(f func(sample)) {
	for e := s.list.Front(); e != nil; e = e.Next() {
		f(e.Value.(sample))
	}
}

func (s *sampleList) forEachPair(f func(sample, sample)) {
	prev := s.list.Front()
	for e := prev.Next(); e != nil; e = e.Next() {
		f(prev.Value.(sample), e.Value.(sample))
		prev = e
	}
}

func absInt64(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

func (s *sampleList) skew() (mean, variance int64) {
	var m, prevM int64 // mean(skew)
	var v int64        // var(skew)
	var n int64
	// Estimate skew using each pair
	s.forEachPair(func(s1, s2 sample) {
		thenInterval := s2.then.Sub(s1.then).Nanoseconds()
		relInterval := s2.rel.sub(s1.rel).Nanoseconds()
		skew := ((thenInterval - relInterval) / relInterval) * 1e9

		prevM = m
		n++
		m = m + (skew-m)/n
		v = v + (skew-m)*(skew-prevM)
	})
	return m, v
}

func (s *sampleList) stats() *statsT {
	r := new(statsT)
	var varSkew int64
	r.skewPPB, varSkew = s.skew()
	stdDevSkew := int64(math.Floor(math.Sqrt(float64(varSkew))))
	r.skewPPBRadius = stdDevSkew * sigmas

	// TODO: Estimate synthetic, using r.shiftAndSkew()
	return r
}
