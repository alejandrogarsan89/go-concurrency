// Package generator shows the "generator" concurrency pattern: a goroutine that
// produces a stream of values on a channel and closes it when done.
//
// Generators are the source stage of a pipeline. Two rules make them safe:
//   - only the sender closes the channel, and it does so exactly once (via
//     defer close), signalling "no more values" to the receiver; and
//   - every generator honours a context.Context so a downstream consumer that
//     stops early can cancel it, preventing the producer goroutine from leaking
//     while blocked on a send.
//
// Each function returns a receive-only channel, so callers cannot accidentally
// close or send on it.
package generator

import "context"

// Ints emits the integers 0, 1, ... n-1 on the returned channel and then closes
// it. If n <= 0 the channel is closed immediately with no values.
//
// The producing goroutine stops early — without leaking — if ctx is cancelled,
// because the send is guarded by a select on ctx.Done().
func Ints(ctx context.Context, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < n; i++ {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// FromSlice emits each element of s in order and then closes the channel. It is
// the canonical way to turn a fixed collection into a stream that feeds a
// pipeline. The producing goroutine exits early if ctx is cancelled.
func FromSlice[T any](ctx context.Context, s []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, v := range s {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Take forwards at most n values from in to the returned channel and then
// closes it, discarding the rest. It is a stream "limiter": it lets a consumer
// read a bounded prefix of a possibly infinite generator and then walk away.
//
// Cancelling ctx stops forwarding early. Take does not drain in; the upstream
// generator should itself observe ctx cancellation to avoid leaking.
func Take[T any](ctx context.Context, in <-chan T, n int) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		if n <= 0 {
			return
		}
		count := 0
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
				count++
				if count == n {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
