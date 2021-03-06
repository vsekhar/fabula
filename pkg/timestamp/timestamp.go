// Package timestamp handles converting time.Time values to unambiguous string
// and byte representations.
//
// Package timestamp handles values roughly between 1678 and 2262, the time
// range that can be represented as nanoseconds from the Unix epoch (1970)
// within an int64. Timestamps for times/dates outside this range are
// undefined.
package timestamp

// NB: time.Time.UnixNano() is undefined for dates outside 1678-2262:
// https://golang.org/pkg/time/#Time.UnixNano

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"
)

var (
	// Earliest is the earliest time that can be converted into a timestamp.
	// Converting times before Earliest into a Timestamp yields undefined
	// behavior. This applies to the zero time.Time{}.
	Earliest = time.Unix(0, math.MinInt64)

	// Latest is the latest time that can be converted into a timestamp.
	// Converting times before Earliest into a Timestamp yields undefined
	// behavior.
	Latest = time.Unix(0, math.MaxInt64)
)

// ToString encodes a time.Time object to a string containing an integer number
// of UTC nanoseconds from the Unix epoch.
func ToString(t time.Time) string {
	n := t.UnixNano()
	return fmt.Sprintf("%d", n)
}

// FromString decodes a time.Time object from a string containing an integer
// number of UTC nanoseconds from the Unix epoch.
func FromString(t string) (time.Time, error) {
	nanos, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, nanos).In(time.UTC), nil
}

// MaxLen is the maximum number of bytes a Timestamp will occupy when converted
// to a binary representation.
const MaxLen = binary.MaxVarintLen64

// ToBytes encodes a time.Time object to a little-endian binary integer number
// of UTC nanoseconds from the Unix epoch, writes the result to a buffer and
// returns the number of bytes written. toBytes panics if the buffer is too
// small.
//
// A variable-length encoding is used; smaller values require fewer bytes. For a
// specification, see
// https://developers.google.com/protocol-buffers/docs/encoding.
func ToBytes(buf []byte, t time.Time) int {
	nanos := t.UnixNano()
	n := binary.PutVarint(buf, nanos) // panics if b is too small
	return n
}

// FromBytes decodes a time.Time object from buf and returns the object and the
// number of bytes read (> 0). If an error occurred, the value is time.Time(0)
// and the number of bytes n is <= 0 with the following meaning:
//
//  n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read
//
// A variable-length encoding is expected; smaller values require fewer bytes.
// For a specification, see
// https://developers.google.com/protocol-buffers/docs/encoding.
func FromBytes(buf []byte) (time.Time, int) {
	nanos, n := binary.Varint(buf)
	t := time.Unix(0, nanos)
	return t, n
}
