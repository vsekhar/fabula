package youtime

import (
	"net"
	"time"
)

const (
	// we have some reason to believe this server knows what time it is.
	serverAuthoritative = iota

	// we are interested in our offsets to this server, but don't think it
	// knows the time better than us.
	serverReporting
)

// compute skew and synthetic sample from at most and at least these numbers of
// samples per server
const maxSamples = 25
const minSamples = 5

var ntpServers = []string{
	// Cloud providers
	"time.google.com:123",   // stratum 1, don't need all four: https://developers.google.com/time/faq#allfour
	"time.facebook.com:123", // stratum 1
	"time.apple.com:123",    // stratum 1
	// "time.cloudflare.com:123", // stratum 3
	// "time.windows.com:123",    // stratum 3

	// Governments
	"time.nist.gov:123", // stratum 1
	"time.nrc.ca:123",   // stratum 2

	// Infrastructure providers
	// "ntp.se:123", // stratum 1, operator of multiple IX's across Sweden
	// "ntp.torix.ca:123", // stratum 1, operator of IX in Toronto, Canada
	// "time.mbix.ca", // stratum ?, operator of IX in Manitoba, Canada

	// "pool.ntp.org:123", // different server every time, noisy?
}

type server struct {
	hostport string
	conn     net.Conn
	samples  sampleList

	// synthetic and its associated radius is the latest estimate of the time
	// at the server and the corresponding relNow. In other words, as of
	// synthetic.rel, our estimate of the time at the server is:
	//
	//   {synthetic.then.Add(-radius), synthetic.then.Add(radius)}.
	//
	// The synthetic sample is adjusted forward (subject to skew correction)
	// to any other relNow to obtain an estimate of the time at the server at a
	// different relNow.
	synthetic sample
	radius    time.Duration

	// skewPPB is how many more ticks synthetic.then ticks for each billion
	// ticks of relNow.
	skewPPB       int64
	skewPPBRadius int64

	err  error
	errN int
}
