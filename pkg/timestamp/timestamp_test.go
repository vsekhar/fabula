package timestamp_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/vsekhar/fabula/pkg/timestamp"
)

var stringCases map[time.Time]string = map[time.Time]string{
	time.Unix(0, 0): "0",
	time.Date(1970, 1, 1, 0, 0, 0, 1, time.UTC): "1",
	time.Unix(0, 500): "500",
	time.Unix(1, 0):   "1000000000",
	time.Date(1970, 1, 1, 0, 0, 1, 1, time.UTC):                          "1000000001",
	time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC):                        "-172800000000000",
	time.Date(1720, 12, 30, 0, 0, 0, 0, time.UTC):                        "-7857820800000000000",
	time.Date(2020, 8, 11, 18, 01, 34, 58982, time.FixedZone("test", 7)): "1597168887000058982",
}

func TestToString(t *testing.T) {
	for v, s := range stringCases {
		if g := timestamp.ToString(v); g != s {
			t.Errorf("bad conversion to string of %s: expected '%s', got '%s'", v, s, g)
		}
	}
}

func TestFromString(t *testing.T) {
	fmt.Printf("epoch: %d", time.Unix(0, 0).Sub(time.Time{}).Nanoseconds())
	for v, s := range stringCases {
		if g, err := timestamp.FromString(s); !g.Equal(v) {
			t.Errorf("bad conversion from string of %s: expected '%s', got '%s' (err: %s)", s, v, g, err)
		}
	}
}

var binaryCases map[time.Time][]byte = map[time.Time][]byte{
	time.Unix(0, 0): {0},
	time.Date(1970, 1, 1, 0, 0, 0, 1, time.UTC): {2},
	time.Unix(0, 500): {232, 7},
	time.Unix(1, 0):   {128, 168, 214, 185, 7},
	time.Date(1970, 1, 1, 0, 0, 1, 1, time.UTC):                          {130, 168, 214, 185, 7},
	time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC):                        {255, 255, 239, 169, 164, 202, 78},
	time.Date(1720, 12, 30, 0, 0, 0, 0, time.UTC):                        {255, 255, 231, 202, 210, 152, 203, 140, 218, 1},
	time.Date(2020, 8, 11, 18, 01, 34, 58982, time.FixedZone("test", 7)): {204, 177, 142, 191, 255, 168, 164, 170, 44},
}

func TestToBytes(t *testing.T) {
	buf := make([]byte, 32)
	for v, b := range binaryCases {
		if n := timestamp.ToBytes(buf, v); !bytes.Equal(buf[:n], b) {
			t.Errorf("bad conversion to bytes of %s: expected '%x', got '%v'", v, b, buf[:n])
		}
	}
}

func TestFromBytes(t *testing.T) {
	for v, b := range binaryCases {
		if g, _ := timestamp.FromBytes(b); !g.Equal(v) {
			t.Errorf("bad conversion from bytes of %v: expected '%s', got '%s'", b, v, g)
		}
	}
}
