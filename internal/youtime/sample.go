package youtime

import "time"

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

type sampleListEntry struct {
	s    sample
	next *sampleListEntry
}

type sampleList struct {
	first *sampleListEntry
	last  *sampleListEntry
	n     int
}

func (s *sampleList) Len() int {
	return s.n
}

func (s *sampleList) AddAndShift(new sample, max int) {
	e := &sampleListEntry{s: new}
	if s.last != nil {
		s.last.next = e
	}
	s.last = e
	if s.first == nil {
		s.first = e
	}
	s.n++
	for s.n > max {
		s.first = s.first.next
		s.n--
	}
}
