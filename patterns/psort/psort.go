// Package psort implements a parallel merge sort: a fork-join sort that uses
// multiple CPU cores to sort large slices faster than a single-threaded sort.
//
// It is a textbook example of divide-and-conquer parallelism. The slice is split
// in half; the two halves are sorted concurrently on separate goroutines and
// then merged. Recursion parallelises only down to a bounded depth (so the
// number of goroutines stays proportional to the number of cores, not the input
// size) and switches to a fast sequential sort for small sub-slices, where the
// overhead of spawning goroutines would outweigh the benefit.
package psort

import (
	"runtime"
	"slices"
	"sync"
)

// seqCutoff is the sub-slice length at or below which we sort sequentially.
// Below it, goroutine and merge overhead dominates any parallel benefit.
const seqCutoff = 1 << 11

// Sort sorts s in place in ascending order according to less, where less(a, b)
// reports whether a must sort before b. It is stable: equal elements keep their
// original relative order. Sorting runs in parallel across the available CPUs
// for large slices and falls back to a sequential sort for small ones.
func Sort[T any](s []T, less func(a, b T) bool) {
	if len(s) <= seqCutoff {
		sequential(s, less)
		return
	}
	buf := make([]T, len(s)) // single scratch buffer, sub-sliced per branch
	parallelSort(s, buf, less, maxDepth())
}

// maxDepth returns how many levels of recursion may fork a goroutine so that the
// number of concurrently sorting goroutines stays around GOMAXPROCS.
func maxDepth() int {
	depth := 0
	for n := runtime.GOMAXPROCS(0); n > 1; n >>= 1 {
		depth++
	}
	return depth
}

func parallelSort[T any](s, buf []T, less func(a, b T) bool, depth int) {
	if len(s) <= seqCutoff || depth <= 0 {
		sequential(s, less)
		return
	}

	mid := len(s) / 2
	// The two halves are disjoint slices (of both s and buf), so sorting them
	// concurrently is race-free.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		parallelSort(s[:mid], buf[:mid], less, depth-1)
	}()
	parallelSort(s[mid:], buf[mid:], less, depth-1)
	wg.Wait()

	merge(s, buf, mid, less)
}

// merge combines the two sorted halves s[:mid] and s[mid:] into a fully sorted s,
// using buf as scratch. The default branch takes the left element on ties, which
// keeps the sort stable.
func merge[T any](s, buf []T, mid int, less func(a, b T) bool) {
	n := len(s)
	copy(buf[:n], s)
	i, j := 0, mid
	for k := 0; k < n; k++ {
		switch {
		case i >= mid:
			s[k] = buf[j]
			j++
		case j >= n:
			s[k] = buf[i]
			i++
		case less(buf[j], buf[i]):
			s[k] = buf[j]
			j++
		default:
			s[k] = buf[i]
			i++
		}
	}
}

// sequential sorts a sub-slice with the standard library's stable sort.
func sequential[T any](s []T, less func(a, b T) bool) {
	slices.SortStableFunc(s, func(a, b T) int {
		switch {
		case less(a, b):
			return -1
		case less(b, a):
			return 1
		default:
			return 0
		}
	})
}
