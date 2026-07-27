// Package pipeline implements the pipeline concurrency pattern: a chain of
// stages connected by channels, where each stage is a goroutine that receives
// values from the previous stage, transforms them, and sends them to the next.
//
// Pipelines decompose a stream-processing task into independent, composable
// stages that run concurrently — while stage 2 processes item N, stage 1 can
// already be producing item N+1. Stages here are ordered (single goroutine each)
// and cancellable; for a parallel stage, use pool.Process.
//
// Because Go generics cannot express a heterogeneous variadic chain, stages are
// composed by nesting calls, e.g.:
//
//	out := pipeline.Filter(ctx,
//	          pipeline.Map(ctx, src, toUpper),
//	          isNonEmpty)
//
// As with any channel pipeline, the caller must ensure the source feeding the
// first stage also observes ctx (all generators in this repo do), otherwise a
// producer can block forever once a cancelled stage stops reading from it.
package pipeline

import "context"

// Map is a pipeline stage that applies fn to each value from in and emits the
// result on the returned channel, preserving order. The output is closed when
// in is drained or ctx is cancelled. Exactly one goroutine runs the stage.
func Map[I, O any](ctx context.Context, in <-chan I, fn func(I) O) <-chan O {
	out := make(chan O)
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- fn(v):
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Filter is a pipeline stage that forwards only the values from in for which
// keep returns true, preserving order. The output is closed when in is drained
// or ctx is cancelled.
func Filter[T any](ctx context.Context, in <-chan T, keep func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				if !keep(v) {
					continue
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
