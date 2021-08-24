package sortablebase64_test

import (
	"testing"

	"github.com/vsekhar/fabula/pkg/sortablebase64"
)

var cases = map[string]uint64{
	"00000000000": 0,
	"00000000001": 1,
	"00000000002": 2,
	"0000000000e": 42,
	"00000000010": 64,
	"0000002usyG": 48992145,
	"Ezzzzzzzzzz": 1<<64 - 1,
}

func TestCases(t *testing.T) {
	for s, n := range cases {
		if ns := sortablebase64.EncodeUint64(n); ns != s {
			t.Errorf("expected %s, got %s", s, ns)
		}
		if sn, err := sortablebase64.DecodeUint64(s); sn != n || err != nil {
			t.Errorf("expected %d, got %d (err: %v)", n, sn, err)
		}
	}
}

func TestSequence(t *testing.T) {
	for i := 0; i < 63; i++ {
		var n uint64 = 1 << i
		ns := sortablebase64.EncodeUint64(n)
		sn, err := sortablebase64.DecodeUint64(ns)
		if sn != n || err != nil {
			t.Errorf("expected %d, got %d (err: %v)", n, sn, err)
		}
	}
}

func TestSortable(t *testing.T) {
	count := 1000
	s := make([]string, count)
	for i := 0; i < count; i++ {
		s[i] = sortablebase64.EncodeUint64(uint64(i))
	}
	for i := range s {
		if i > 0 {
			if !(s[i-1] < s[i]) {
				t.Errorf("bad ordering: %s >= %s", s[i-1], s[i])
			}
			if next, err := sortablebase64.IncUint64(s[i-1]); next != s[i] {
				t.Errorf("failed to increment %s to %s, got %s (err: %s)", s[i-1], s[i], next, err)
			}
			if prev, err := sortablebase64.DecUint64(s[i]); prev != s[i-1] || err != nil {
				t.Errorf("failed to decrement %s to %s, got %s (err: %s)", s[i], s[i-1], prev, err)
			}
		}
	}
}
