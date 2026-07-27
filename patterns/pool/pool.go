// Package pool implements the worker-pool pattern: a fixed number of worker
// goroutines that share the work of processing many items.
//
// A worker pool is how you get controlled parallelism. Spawning one goroutine
// per item (as waitgroup.Map does) is fine for a small, bounded set, but for
// thousands of items — especially CPU-bound or resource-hungry work — you want
// to cap concurrency at a chosen number of workers. The pool "fans out" jobs to
// the workers and "fans in" their results, all cancellable via context.
package pool

import (
	"context"
	"sync"
)

// Process fans out the values from in across the given number of worker
// goroutines, applies fn to each, and returns a channel of results. Results are
// interleaved in completion order (unordered); use Map when you need results in
// input order.
//
// The output channel is closed once every worker has exited, which happens when
// in is drained or ctx is cancelled. workers is clamped to at least 1. The
// caller is responsible for cancelling the producer that feeds in (typically by
// sharing ctx) so it does not block after the pool stops reading.
func Process[T, R any](ctx context.Context, in <-chan T, workers int, fn func(context.Context, T) R) <-chan R {
	if workers < 1 {
		workers = 1
	}
	out := make(chan R)

	var wg sync.WaitGroup
	wg.Add(workers)
	worker := func() {
		defer wg.Done()
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return // input drained
				}
				select {
				case out <- fn(ctx, v):
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}

	for i := 0; i < workers; i++ {
		go worker()
	}

	// One closer goroutine closes out exactly once, after all workers exit.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Map applies fn to every element of inputs using at most workers goroutines and
// returns the results in the same order as the inputs. Unlike Process it is not
// streaming: it blocks until all inputs are processed and returns a slice.
//
// Concurrency is bounded by workers (clamped to at least 1), so it is safe for
// large inputs. Each result is written to a distinct index, so no locking is
// needed. If ctx is cancelled mid-flight, dispatch stops and the entries that
// were never processed keep their zero value; fn should observe ctx for
// long-running work.
func Map[T, R any](ctx context.Context, inputs []T, workers int, fn func(context.Context, T) R) []R {
	results := make([]R, len(inputs))
	if len(inputs) == 0 {
		return results
	}
	if workers < 1 {
		workers = 1
	}

	type job struct {
		idx int
		val T
	}
	jobs := make(chan job)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.idx] = fn(ctx, j.val) // distinct index => race-free
			}
		}()
	}

	// Dispatch jobs; stop early if the context is cancelled. Closing jobs lets
	// the workers drain and exit.
	go func() {
		defer close(jobs)
		for i, v := range inputs {
			select {
			case jobs <- job{idx: i, val: v}:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	return results
}
