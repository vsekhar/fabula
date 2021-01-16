package youtime

import (
	"net"
	"testing"
)

func TestCodedProbes(t *testing.T) {
	nc, err := net.Dial("udp", "time.google.com:123")
	if err != nil {
		t.Fatal(err)
	}
	nonPure := 0
	for i := 0; i < 10; i++ {
		out, in, err := GetCodedProbes(nc)
		if err != nil {
			if err == ErrCodedProbesNotPure {
				nonPure++
			}
			t.Error(err)
		}
		t.Logf("jitter: out %s, in %s", Jitter(out[0], out[1]), Jitter(in[0], in[1]))
	}
	if nonPure > 0 {
		t.Errorf("Non-pure probes: %d", nonPure)
	}
	t.Error(1)
}
