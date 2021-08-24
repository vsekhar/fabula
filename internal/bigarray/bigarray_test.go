package bigarray_test

import (
	"testing"

	"github.com/vsekhar/fabula/internal/bigarray"
)

var cases [][]int = [][]int{
	// next, min, expectedProbes
	{0, 0, 1},
	{0, 6, 1},
	{100, 1, 16},
	{4096, 100, 26},
	{1e6, 0, 42},
	{1e9, 0, 62},
	{9999999999, 0, 70},
	{9999999999, 9999999999, 1},
}

func TestSearch(t *testing.T) {
	for _, c := range cases {
		expectedNext, min, expectedProbes := c[0], c[1], c[2]
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

func TestSearchBatch(t *testing.T) {
	for _, c := range cases {
		expectedNext, min, _ := c[0], c[1], c[2]
		probes := 0
		batchProbe := func(i int) (bool, int) {
			probes++
			// pretend like we probe all the way up to expectedNext
			return true, expectedNext
		}
		next := bigarray.SearchBatch(min, batchProbe)
		if expectedNext < min {
			expectedNext = min
		}
		if next != expectedNext || probes != 1 {
			t.Errorf("expected %d (%d probes), got %d (%d probes)", expectedNext, 1, next, probes)
		}
	}
}
