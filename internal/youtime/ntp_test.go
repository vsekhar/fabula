package youtime

import (
	"testing"
	"time"
)

func timeOrPanic(format, s string) time.Time {
	t, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestTimeConversions(t *testing.T) {
	cases := []time.Time{
		time.Now(),
		time.Unix(42, 84),
		timeOrPanic(time.RFC3339, "2020-04-22T17:22:09-07:00"),
		timeOrPanic(time.RFC3339Nano, "2006-01-02T15:04:05.123456789-07:00"),
	}
	for _, c := range cases {
		s, f := timeToNTP(c)
		nc := ntpToTime(s, f)
		if !nc.Equal(c) {
			t.Errorf("mismatch: ntpToTime(%d, %d)==%s, expected: %s", s, f, nc, c)
		}
	}
}
