// +build linux

// Package youtime provides bounded time uncertainty.
//
// YouTime is roughly inspired by Geng, et al., "Exploitng a Natural Network Effect
// for Scalable Fine-grained Clock Synchronization," Proc. 15th USENIX Sym. on
// NSDI, 2018.
//
// https://www.usenix.org/system/files/conference/nsdi18/nsdi18-geng.pdf
package youtime

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

const startingRadius = 5 * time.Second

// TODO:
//
// 1) estimate skew by looking at skew between samples from a server. Each
// interval provides a sample of the skew ("I ticked X, you ticked Y"). Measured
// ppb, skew = (y2-y1)/(x2-x1) * 1e9. Assign that sample of skew to the relNow
// at the midpoint of the interval.
//
// 2) Use an exponential moving average of skew to compute latest best estimate
// of local to server skew. More recent samples of skew are weighted higher.
//
// 3) Obtain a local relNow
//
// 4) Obtain expected now from each server sample, adjusted for that sample's
// skew, for that relNow. Possibly weight these exponentially. This is the
// now of an "ideal clock" corresponding to local relNow from step #3. This
// now and the relNow from #3 constitute a synthetic sample of this ideal
// clock, against which we will serve time locally.
//
// 5) System skew to this ideal clock is the average of skews to each server.
//
// 6) Time can now be served using relative intervals from the synthetic sample
// to the time request moment. Skew is adjusted using skew to the ideal clock.
//
// 7) Uncertainty depends on the stdev around skew calculations and the stdev
// of the time of the ideal clock.
//
// As a simple version, skip the frequency step, calculate expected now for a
// relNow using all samples. The range on these expected nows are your
// uncertainty bounds. Check if frequency-adjusting samples improves or reduces
// uncertainty.

type statsT struct {
	ideal      sample        // synthetic sample of an "ideal" clock
	radius     time.Duration // radius around now in synthetic sample (TODO: stdev?)
	skew       int64         // estimated skew from relNow to synthetic sample in ppb
	skewRadius int64         // radius around skew (TODO: stdev?)
}

// Client is an instance of YouTime.
type Client struct {
	ctx       context.Context
	stats     atomic.Value // statsT
	servers   map[string]*server
	readyCh   chan struct{} // closed when first stats are published
	readyOnce sync.Once
}

// NewClient returns a new YouTime client.
func NewClient(ctx context.Context) *Client {
	c := &Client{
		ctx:     ctx,
		servers: make(map[string]*server),
		readyCh: make(chan struct{}),
	}
	ticker := time.NewTicker(updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.update(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return c
}

func (c *Client) loadStats() *statsT {
	return c.stats.Load().(*statsT)
}

func (c *Client) storeStats(s *statsT) {
	c.stats.Store(s)
}

func getSample(nc net.Conn) (sample, error) {
	req := &packet{Settings: 0x1B}
	rsp := &packet{}
	s := relNow()
	if err := binary.Write(nc, binary.BigEndian, req); err != nil {
		return sample{}, err
	}
	if err := binary.Read(nc, binary.BigEndian, rsp); err != nil {
		return sample{}, err
	}
	m := mid(s, relNow())
	secs := float64(rsp.RxTimeSec) - ntpEpochOffset
	nanos := (int64(rsp.RxTimeFrac) * 1e9) >> 32
	t := time.Unix(int64(secs), nanos)
	return sample{then: t, rel: m}, nil
}

var errCodedProbesNotPure = errors.New("coded probes not pure")

// CodedProbeInterval is the time between coded probes.
const CodedProbeInterval = 1 * time.Second
const codedProbeEpsilon = 5 * time.Millisecond
const maxServerError = 3 // consecutive errors
const updateInterval = 2 * time.Second

func getCodedSamples(nc net.Conn) (s1, s2 sample, err error) {
	dl := time.Now().Add(ntpTimeout)
	if err = nc.SetDeadline(dl); err != nil {
		return
	}
	s1, err = getSample(nc)
	if err != nil {
		return
	}
	time.Sleep(CodedProbeInterval)
	dl = time.Now().Add(ntpTimeout)
	if err = nc.SetDeadline(dl); err != nil {
		return
	}
	s2, err = getSample(nc)
	if err != nil {
		return
	}
	if s1.then.After(s2.then) {
		return sample{}, sample{}, errCodedProbesNotPure
	}
	jitter := s2.then.Sub(s1.then) - s2.rel.sub(s1.rel)
	if jitter < 0 {
		jitter = -jitter
	}
	if jitter >= codedProbeEpsilon {
		return sample{}, sample{}, errCodedProbesNotPure
	}
	return s1, s2, nil
}

func (c *Client) fetchSamples(ctx context.Context) error {
	eg, _ := errgroup.WithContext(ctx)
	// fetch a set of samples for each server (adding it if it doesn't exist)
	for _, hp := range ntpServers {
		srv, ok := c.servers[hp]
		if !ok {
			// create server entry
			nc, err := net.Dial("udp", hp)
			if err != nil {
				return err
			}
			srv = &server{}
			srv.hostport = hp
			srv.conn = nc
			c.servers[hp] = srv
		}

		eg.Go(func() error {
			var err error
			for {
				var s1, s2 sample
				s1, s2, err = getCodedSamples(srv.conn)
				if err != nil {
					log.Print(err)
					srv.err = err
					srv.errN++
					if srv.errN > maxServerError {
						return fmt.Errorf("too many errors from %s: %s", srv.hostport, err)
					}
					continue
				}
				srv.err = nil
				srv.errN = 0
				srv.samples.AddAndShift(s1, maxSamples)
				srv.samples.AddAndShift(s2, maxSamples)
				break
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Client) update(ctx context.Context) error {
	if err := c.fetchSamples(ctx); err != nil {
		return err
	}
	// TODO: for each server, compute average skew between pairs of samples, and
	// skew radius. Use skew to compute ideal sample and radius.
	// Create a new statsT

	fakeStats := &statsT{ideal: sample{rel: relNow(), then: time.Now()}}
	c.stats.Store(fakeStats)
	c.readyOnce.Do(func() {
		close(c.readyCh)
	})
	return nil
}

// Ready blocks until the client is ready.
func (c *Client) Ready() {
	<-c.readyCh
}

// Uncertainty returns the current estimated uncertainty for the YouTimes
// produced by Get.
func (c *Client) Uncertainty() time.Duration {
	return c.get().EstimatedCommitWait()
}

func (c *Client) getRange() (earliest, latest time.Time) {
	s := c.loadStats()
	if s == nil {
		panic("called Get before calling Start")
	}
	rnow := relNow()
	mid := s.ideal.then.Add(s.ideal.rel.to(rnow))

	lowSkew := (s.ideal.rel.to(rnow).Nanoseconds() / 1e9) * (s.skew - s.skewRadius)
	edelta := -s.radius + time.Duration(lowSkew)

	highSkew := (s.ideal.rel.to(rnow).Nanoseconds() / 1e9) * (s.skew + s.skewRadius)
	ldelta := s.radius + time.Duration(highSkew)

	return mid.Add(edelta), mid.Add(ldelta)
}

// YouTime represents a range of time within which the current time lies.
type YouTime struct {
	client   *Client
	earliest time.Time
	latest   time.Time
}

// EstimatedCommitWait returns the maximum amount of time Time would wait
// before providing a timestamp.
func (y YouTime) EstimatedCommitWait() time.Duration {
	return time.Duration(y.latest.Sub(y.earliest).Nanoseconds() / 2)
}

// Time provides the current time. Time blocks until it can be sure the current
// time is in the past. That is, after Now returns, no well-synced instance of
// YouTime will return an earlier time.
func (y *YouTime) Time() time.Time {
	ny := y.client.get()
	for {
		if ny.earliest.After(y.latest) {
			break
		}
		time.Sleep(y.latest.Sub(ny.earliest))
		ny = y.client.get()
	}
	return y.latest
}

// Get returns a YouTime range.
//
// Other work can be done after calling Get and before extracting the timestamp
// by calling Time on the returned YouTime.
func (c *Client) get() YouTime {
	earliest, latest := c.getRange()
	return YouTime{
		client:   c,
		earliest: earliest,
		latest:   latest,
	}
}

// TODO: YouTime instances sync with each other, not to improve their time
// estimates (they only use authoritative servers for that), but to check and
// report on the offsets and uncertainties they see. In particular, they should
// confirm that their current commit wait window is long enough to wait out any
// uncertainty they are seeing from other YouTime instances.
//
// YouTime as NTP server? On a heigher port? UDP is ideal.
//
// Of particular interest is how synchronized YouTime servers can get across
// clouds.
