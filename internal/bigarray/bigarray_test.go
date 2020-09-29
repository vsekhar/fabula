package bigarray_test

import (
	"testing"

	"github.com/vsekhar/fabula/internal/bigarray"
)

var cases [][]int = [][]int{
	// next, min, expectedProbes
	{0, 0, 2},
	{0, 6, 2},
	{100, 1, 15},
	{4096, 100, 25},
	{9999999999, 0, 69},
	{9999999999, 9999999999, 2},
}

func TestSearch(t *testing.T) {
	for _, c := range cases {
		expectedNext := c[0]
		min := c[1]
		expectedProbes := c[2]
		probes := 0
		probe := func(i int) bool {
			probes++
			return i >= expectedNext
		}
		next := bigarray.Search(min, probe)
		if expectedNext < min {
			expectedNext = min
		}
		if next != expectedNext || probes != expectedProbes {
			t.Errorf("expected %d (%d probes), got %d (%d probes)", expectedNext, expectedProbes, next, probes)
		}
	}
}
