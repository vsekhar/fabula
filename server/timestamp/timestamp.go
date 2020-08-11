// Package timestamp handles converting time.Time values to unambiguous string
// and byte representations.
package timestamp

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
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
