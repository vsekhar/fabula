package youtime

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"
)

// TODO: make logger compliant with traceability:
//   https://www.industry.gov.au/sites/default/files/2019-11/nmi-using-ntp-for-traceable-time-and-frequency.pdf

// Anything more than this and we will likely fail the coded probe jitter tests.
const ntpTimeout = 100 * time.Millisecond

const maxServerError = 3 // consecutive errors
const updateInterval = 2 * time.Second

type source struct {
	hosts  []string
	logger *log.Logger

	conn    net.Conn
	samples sampleList
	stats   atomic.Value // statsT

	err  error
	errN int
}

func (s *source) setStats(stats *statsT) {
	s.stats.Store(stats)
}

func (s *source) getStats() *statsT {
	return s.stats.Load().(*statsT)
}

func (s *source) ready() bool {
	return s.getStats() != nil
}

func (s *source) log(v ...interface{}) {
	if s.logger != nil {
		s.logger.Output(2, fmt.Sprint(v...))
	}
}

func (s *source) logf(f string, vars ...interface{}) {
	if s.logger != nil {
		s.logger.Output(2, fmt.Sprintf(f, vars...))
	}
}
