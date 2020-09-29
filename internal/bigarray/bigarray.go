// Package bigarray contains functions for working with very large (effectively
// infinite) arrays.
package bigarray

import (
	"sort"
)

// Search uses binary search to find and return the smallest index i in
// [min, maxInt] at which f(i) is true, assuming that on the range
// [min, maxInt], f(i) == true implies f(i+1) == true. That is, Search requires
// that f is false some (possibly empty) prefix of the input range [min, maxInt]
// and then true thereafter.
func Search(min int, f func(int) bool) int {
	// Shift the search from [min, maxInt] to [0, maxInt-min]
	g := func(i int) bool {
		return f(min + i)
	}

	// Exponentially find a reasonable upper bound.
	high := 1
	for !g(high) {
		high *= 2
	}

	return sort.Search(high, g) + min
}
