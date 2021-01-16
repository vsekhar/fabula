package youtime

import (
	"math"
	"time"
)

// https://medium.com/learning-the-go-programming-language/lets-make-an-ntp-client-in-go-287c4b9a969f

type packet struct {
	Settings       uint8  // leap yr indicator, ver number, and mode
	Stratum        uint8  // stratum of local clock
	Poll           int8   // poll exponent
	Precision      int8   // precision exponent
	RootDelay      uint32 // root delay
	RootDispersion uint32 // root dispersion
	ReferenceID    uint32 // reference id
	RefTimeSec     uint32 // reference timestamp sec
	RefTimeFrac    uint32 // reference timestamp fractional
	OrigTimeSec    uint32 // origin time secs
	OrigTimeFrac   uint32 // origin time fractional
	RxTimeSec      uint32 // receive time secs
	RxTimeFrac     uint32 // receive time frac
	TxTimeSec      uint32 // transmit time secs
	TxTimeFrac     uint32 // transmit time frac
}

const ntpEpochOffset = 2208988800 // seconds

func ntpToTime(sec, frac uint32) time.Time {
	// secs := float64(rsp.RxTimeSec) - ntpEpochOffset
	secs := int64(sec) - ntpEpochOffset
	nanos := (int64(frac) * 1e9) >> 32
	return time.Unix(secs, nanos)
}

func timeToNTP(t time.Time) (sec, frac uint32) {
	secs := t.Unix() + ntpEpochOffset
	nanos := int64(t.Nanosecond())
	nanos = (nanos << 32) / 1e9
	if secs > math.MaxUint32 {
		panic("overflow")
	}
	return uint32(secs), uint32(nanos) + 1 // because of course
}
