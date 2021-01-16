package youtime

/*
#cgo LDFLAGS: -L${SRCDIR}/../../third_party/github.com/cjlin1/liblinear
#include "../../third_party/github.com/cjlin1/liblinear/linear.h"
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"net"
	"time"

	"gonum.org/v1/gonum/mat"
)

// ErrCodedProbesNotPure is returned when a pair of probes have jitter greater
// than CodedProbeEpsilon.
var ErrCodedProbesNotPure = errors.New("youtime/probe: coded probes not pure")

// CodedProbeInterval is the time between coded probes.
const CodedProbeInterval = 1 * time.Second

// CodedProbeEpsilon is the maximum jitter between coded probe pairs
//
// TODO: experiment to see if bringing this down reduces uncertainty (at the
// cost of settling time).
//
// As of Oct 9 from TekSavvy@TOR+WLAN, ~35% pass @ 1ms, ~70% @ 2ms.
const CodedProbeEpsilon = 2 * time.Millisecond

// Probe represents a single probe of a remote time source.
//
// A probe can be obtained in either the inbound or outbound direction (the
// probe itself does not record its direction).
type Probe struct {
	// Local is the time the packet left the local host (if outbound) or the
	// time the packet arrived at the local host (if inbound).
	local relMoment

	// Remote is the time the packet arrived at the remote host (if outbound) or
	// the time the packet left the remote host (if inbound).
	remote time.Time
}

func probesToSample(out, in Probe) sample {
	return sample{
		rel:  mid(in.local, out.local),
		then: out.remote.Add(in.remote.Sub(out.remote) / 2),
	}
}

// Bound returns the bound estimate of clock discrepancy based on the probe. The
// return value is an upper bound if p is an outbound probe, and a lower bound
// if p is an inbound probe.
func Bound(p Probe) int64 {
	return p.remote.UnixNano() - p.local.rel
}

// DebugLocal returns the reading from the local relative clock for a given
// probe. It should only be used for tests and generating debug output.
//
// The local relative clock reading may have been taken before the probe's
// remote reading (in the case of an outbound probe) or after the probe's remote
// reading (in the case of an inbound probe).
func DebugLocal(p Probe) int64 {
	return p.local.rel
}

// Jitter returns the difference in discrepancy between two probes, a measure
// of network and host uncertainties.
func Jitter(p1, p2 Probe) time.Duration {
	j := p2.remote.Sub(p1.remote) - p2.local.sub(p1.local)
	if j < 0 {
		return -j
	}
	return j
}

// GetProbe performs a single probe over nc.
func GetProbe(nc net.Conn) (out, in Probe, err error) {
	osec, ofrac := timeToNTP(time.Now()) // used as a nonce, not for time sync
	req := &packet{
		Settings:   0x1B,
		TxTimeSec:  osec,
		TxTimeFrac: ofrac,
	}
	rsp := &packet{}
	{
		// Critical timing block
		out.local = relNow()
		if err = binary.Write(nc, binary.BigEndian, req); err != nil {
			return
		}
	}
	for {
		{
			// Critical timing block
			if err = binary.Read(nc, binary.BigEndian, rsp); err != nil {
				return
			}
			in.local = relNow()
		}
		if rsp.OrigTimeSec == osec && rsp.OrigTimeFrac == ofrac {
			break
		}
		// not our packet, read again (or timeout)
	}
	out.remote = ntpToTime(rsp.RxTimeSec, rsp.RxTimeFrac)
	in.remote = ntpToTime(rsp.TxTimeSec, rsp.TxTimeFrac)
	return
}

// GetCodedProbes returns a set of outbound and inbound probes, or an error.
func GetCodedProbes(nc net.Conn) (outProbes, inProbes []Probe, err error) {
	const N = 2
	for i := 0; i < N; i++ {
		if i != 0 {
			time.Sleep(CodedProbeInterval)
		}
		dl := time.Now().Add(ntpTimeout)
		if err = nc.SetDeadline(dl); err != nil {
			return nil, nil, err
		}
		out, in, err := GetProbe(nc)
		if err != nil {
			return nil, nil, err
		}
		outProbes = append(outProbes, out)
		inProbes = append(inProbes, in)
	}

	if outProbes[0].remote.After(outProbes[1].remote) {
		return outProbes, inProbes, ErrCodedProbesNotPure
	}
	if inProbes[0].remote.After(inProbes[1].remote) {
		return outProbes, inProbes, ErrCodedProbesNotPure
	}
	if j := Jitter(outProbes[0], outProbes[1]); j >= CodedProbeEpsilon {
		return outProbes, inProbes, ErrCodedProbesNotPure
	}
	if j := Jitter(inProbes[0], inProbes[1]); j >= CodedProbeEpsilon {
		return outProbes, inProbes, ErrCodedProbesNotPure
	}
	return
}

// ProbeStore is a dual circular buffer for storing and reading upper and lower
// bound probes.
type ProbeStore struct {
	out    []Probe
	in     []Probe
	max    int
	offset int
}

// NewProbeStore returns a new ProbeStore that will store at most max outbound
// and max inbound probes.
func NewProbeStore(max int) *ProbeStore {
	return &ProbeStore{max: max}
}

// Len returns the number of outbound and inbound pairs of probes in the store.
func (ps *ProbeStore) Len() int {
	return len(ps.out)
}

// Add adds a new pair of probes, out and in, to the store.
func (ps *ProbeStore) Add(out, in Probe) {
	if len(ps.out) < ps.max {
		ps.out = append(ps.out, out)
		ps.in = append(ps.in, in)
	} else {
		ps.out[ps.offset] = out
		ps.in[ps.offset] = in
		ps.offset++
	}
}

func (ps *ProbeStore) xMatrix() mat.Matrix {
	r := &xMatrixT{}
	r.ps = ps
	return r
}

func (ps *ProbeStore) yMatrix() mat.Matrix {
	r := &yMatrixT{}
	r.ps = ps
	return r
}

type xMatrixT struct {
	ps *ProbeStore
}

// var _ mat.RowViewer = (*xMatrixT)(nil)
// var _ mat.RowViewer = (*yMatrixT)(nil)

func (x *xMatrixT) At(i, j int) float64 {
	var e *Probe
	switch {
	case i < len(x.ps.out):
		e = &x.ps.out[i]
	case i-len(x.ps.out) < len(x.ps.in):
		e = &x.ps.in[i-len(x.ps.out)]
	default:
		panic("out of range")
	}
	switch j {
	case 0:
		return float64(e.local.rel)
	case 1:
		return float64(e.remote.UnixNano())
	}
	panic("out of range")
}

func (x *xMatrixT) T() mat.Matrix {
	return &mat.Transpose{Matrix: x}
}

func (x *xMatrixT) Dims() (r, c int) {
	return len(x.ps.out) + len(x.ps.in), 2
}

type yMatrixT struct {
	ps *ProbeStore
}

func (y *yMatrixT) At(i, j int) float64 {
	if j != 0 {
		panic("out of range")
	}
	switch {
	case i < len(y.ps.out):
		return 1.0
	case i-len(y.ps.out) < len(y.ps.in):
		return -1.0
	default:
		panic("out of range")
	}
}

func (y *yMatrixT) T() mat.Matrix {
	return &mat.Transpose{Matrix: y}
}

func (y *yMatrixT) Dims() (r, c int) {
	return len(y.ps.out) + len(y.ps.in), 1
}

func (ps *ProbeStore) asCArrays() (x **C.struct_feature_node, y *C.double) {
	// TODO: allocate with C.malloc and copy values sorted with positive
	// examples first ([]out) and negative examples second ([]in).

	// TODO: return struct that handles freeing this memory and avoids leaking
	// the use of cgo? Or just do all the SVM stuff internal to the library
	// and have ProbeStore provide learned coefficients in terms of time: skew
	// and offset.

	panic("unimplemented")
}
