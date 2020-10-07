package sortablebase64_test

import (
	"testing"

	"github.com/vsekhar/fabula/pkg/sortablebase64"
)

var cases = map[string]uint64{
	"-----------": 0,
	"----------0": 1,
	"----------1": 2,
	"----------e": 42,
	"---------0-": 64,
	"------1usyG": 48992145,
	"Ezzzzzzzzzz": 1<<64 - 1,
}

func TestCases(t *testing.T) {
	for s, n := range cases {
		if ns := sortablebase64.EncodeUint64(n); ns != s {
			t.Errorf("expected %s, got %s", s, ns)
		}
		if sn := sortablebase64.DecodeUint64(s); sn != n {
			t.Errorf("expected %d, got %d", n, sn)
		}
	}
}

func TestSequence(t *testing.T) {
	for i := 0; i < 63; i++ {
		var n uint64 = 1 << i
		ns := sortablebase64.EncodeUint64(n)
		sn := sortablebase64.DecodeUint64(ns)
		if sn != n {
			t.Errorf("expected %d, got %d", n, sn)
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
			if next := sortablebase64.Inc(s[i-1]); next != s[i] {
				t.Errorf("failed to increment %s to %s, got %s", s[i-1], s[i], next)
			}
			if prev := sortablebase64.Dec(s[i]); prev != s[i-1] {
				t.Errorf("failed to decrement %s to %s, got %s", s[i], s[i-1], prev)
			}
		}
	}
}
