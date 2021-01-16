package youtime

import (
	"time"

	"golang.org/x/sys/unix"
)

type relMoment struct {
	// this is a struct instead of a straight int64 because we don't want
	// the value in here to be accidentally used for anything other than
	// relative calculations
	rel int64
}

// relNow returns a value representing the current moment in time relative to
// some other value from relNow.
func relNow() relMoment {
	var t unix.Timespec
	err := unix.ClockGettime(unix.CLOCK_MONOTONIC_RAW, &t)
	if err != nil {
		panic(err)
	}
	return relMoment{rel: t.Nano()}
}

// sub returns the Duration from s to r.
//
// sub is equivalent to calling s.to(r).
func (r relMoment) sub(s relMoment) time.Duration {
	return time.Duration(r.rel - s.rel)
}

// To returns the Duration from r to s.
//
// To is equivalent to calling s.sub(r).
func (r relMoment) to(s relMoment) time.Duration {
	return time.Duration(s.rel - r.rel)
}

// Mid returns the relMoment midway between r and s.
func mid(r, s relMoment) relMoment {
	return relMoment{rel: r.rel + ((s.rel - r.rel) / 2)}
}
