// Package mapreduce implements a parallel map-reduce over a slice: map a
// function across every element and fold the results into a single value, using
// multiple CPU cores.
//
// This is the canonical way to turn a CPU-bound aggregation into real
// parallelism. The input is split into contiguous chunks, one per worker; each
// worker maps and reduces its chunk into a partial result on its own goroutine
// (writing a distinct slice index, so no locking is needed), and the partials
// are combined at the end. Because the chunks are contiguous and combined in
// order, the result equals a sequential left fold as long as the reducer is
// associative with identity as its neutral element.
package mapreduce

import (
	"context"
	"sync"
)

// MapReduce applies mapper to every element of inputs and folds the mapped
// values with reducer, in parallel across at most workers goroutines. identity
// is the neutral element of reducer (reducer(identity, x) == x); it is also the
// result for empty input.
//
// reducer must be associative — reducer(a, reducer(b, c)) == reducer(reducer(a,
// b), c) — because the fold is computed per chunk and then across chunks. It
// need not be commutative: chunks are contiguous and combined left to right, so
// the result matches the equivalent sequential fold.
//
// workers is clamped to at least 1 and never exceeds len(inputs). If ctx is
// cancelled mid-flight, each worker stops at the next element and MapReduce
// returns the fold of whatever was processed so far (a best-effort partial).
func MapReduce[In, Out any](
	ctx context.Context,
	inputs []In,
	workers int,
	mapper func(In) Out,
	reducer func(a, b Out) Out,
	identity Out,
) Out {
	n := len(inputs)
	if n == 0 {
		return identity
	}
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}

	partials := make([]Out, workers)
	chunk := (n + workers - 1) / workers // ceil division

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		start := w * chunk
		if start >= n {
			partials[w] = identity
			continue
		}
		end := start + chunk
		if end > n {
			end = n
		}

		wg.Add(1)
		go func(w, start, end int) {
			defer wg.Done()
			acc := identity
			for i := start; i < end; i++ {
				if ctx.Err() != nil { // cheap cooperative cancellation
					break
				}
				acc = reducer(acc, mapper(inputs[i]))
			}
			partials[w] = acc // distinct index => race-free
		}(w, start, end)
	}
	wg.Wait()

	result := identity
	for _, p := range partials {
		result = reducer(result, p)
	}
	return result
}
