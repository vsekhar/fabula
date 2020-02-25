// Package timestamp handles converting Timestamp values to canonical string
// and byte representations.
package timestamp

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
)

// toString encodes a time.Time object to a string containing an integer number
// of UTC nanoseconds from the Unix epoch.
func toString(t time.Time) string {
	return fmt.Sprintf("%d", t.UnixNano())
}

// fromString decodes a time.Time object from a string containing an integer
// number of UTC nanoseconds from the Unix epoch.
func fromString(t string) (time.Time, error) {
	nanos, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, nanos), nil
}

// MaxLen is the maximum number of bytes a Timestamp will occupy when converted
// to a binary representation.
const MaxLen = binary.MaxVarintLen64

// toBytes encodes a time.Time object to a little-endian binary integer number
// of UTC nanoseconds from the Unix epoch, writes the result to a buffer and
// returns the number of bytes written. toBytes panics if the buffer is too
// small.
//
// A variable-length encoding is used; smaller values require fewer bytes. For a
// specification, see
// https://developers.google.com/protocol-buffers/docs/encoding.
func toBytes(buf []byte, t time.Time) int {
	nanos := t.UnixNano()
	n := binary.PutVarint(buf, nanos) // panics if b is too small
	return n
}

// fromBytes decodes a time.Time object from buf and returns the object and the
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
func fromBytes(buf []byte) (time.Time, int) {
	nanos, n := binary.Varint(b)
	t := time.Unix(0, nanos)
	return t, n
}
