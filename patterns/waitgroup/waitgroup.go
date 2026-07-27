// Package waitgroup demonstrates the most fundamental concurrency primitive in
// Go: launching goroutines and waiting for them to finish with a sync.WaitGroup.
//
// A WaitGroup is a counter. Add sets how many goroutines to wait for, each
// goroutine calls Done when it finishes (idiomatically via defer), and Wait
// blocks until the counter reaches zero. Getting this pattern right — always
// Add before starting the goroutine, always Done exactly once — is the basis of
// almost every higher-level pattern in this library.
package waitgroup

import "sync"

// RunAll runs every task in its own goroutine and blocks until all of them have
// returned. It spawns one goroutine per task, so it is meant for a bounded,
// known set of tasks; for large or unbounded workloads use a worker pool
// instead (see the pool pattern).
//
// A nil task is skipped rather than causing a panic, so callers may pass a
// sparse list safely.
func RunAll(tasks ...func()) {
	var wg sync.WaitGroup
	for _, task := range tasks {
		if task == nil {
			continue
		}
		wg.Add(1)
		go func(fn func()) {
			defer wg.Done() // Done even if fn panics, so Wait never hangs.
			fn()
		}(task)
	}
	wg.Wait()
}

// Map applies fn to every element of inputs concurrently and returns the
// results in the same order as the inputs. Each call to fn runs in its own
// goroutine; results are written to distinct slice indices, so no locking is
// needed even though the goroutines run in parallel.
//
// fn must be safe to call concurrently. For CPU-bound work the effective
// parallelism is capped by GOMAXPROCS. Order is preserved regardless of the
// order in which the goroutines complete.
func Map[T, R any](inputs []T, fn func(T) R) []R {
	results := make([]R, len(inputs))
	var wg sync.WaitGroup
	wg.Add(len(inputs))
	for i, in := range inputs {
		go func(idx int, v T) {
			defer wg.Done()
			results[idx] = fn(v) // distinct index => data-race free
		}(i, in)
	}
	wg.Wait()
	return results
}
