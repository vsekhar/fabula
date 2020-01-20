package main

import (
	"fmt"
	"strconv"
	"time"
)

// toString encodes a time.Time object as an integer number of UTC nanoseconds
// from the Unix epoch.
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
