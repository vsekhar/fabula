// Package bigarray contains functions for working with very large (effectively
// infinite) arrays.
package bigarray

import (
	"sort"
)

// Search uses binary search to find and return the smallest index i in
// [min, maxInt] at which f(i) is true, assuming that on the range
// [min, maxInt], f(i) == true implies f(i+1) == true. That is, Search requires
// that f is false for some (possibly empty) prefix of the input range [min,
// maxInt] and then true thereafter.
func Search(min int, f func(i int) (atI bool)) int {
	return SearchBatch(min, func(i int) (bool, int) {
		return f(i), i
	})
}

// SearchBatch is like Search, but f can return additional information to
// improve performance. SearchBatch expects f to probe at index i and
// sequentially zero or more additional indices > i, and return upon finding the
// first index at which its condition evaluates to true.
//
// atLastChecked should be true if the condition is true at index lastChecked.
//
// lastChecked >= i is the last index SearchBatch checked.
func SearchBatch(min int, f func(i int) (atLastChecked bool, lastChecked int)) int {
	// Shift the search from [min, maxInt] to [0, maxInt-min]
	g := func(i int) (bool, int) {
		t, lastChecked := f(min + i)
		return t, lastChecked - min
	}

	// Exponentially find a reasonable upper bound.
	high := 0
	for {
		t, lastChecked := g(high)
		if t {
			// threshold is <= lastChecked, we found the value
			if lastChecked > high {
				return lastChecked + min
			}
			// threshold is < high, we're done finding an upper bound, jump to
			// binary search between [0,high].
			break
		}
		if high == 0 {
			high = 1
		} else {
			high *= 2
		}
		if (lastChecked + 1) > high {
			high = lastChecked + 1
		}
	}

	return sort.Search(high, func(i int) bool { r, _ := g(i); return r }) + min
}
