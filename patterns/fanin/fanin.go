// Package fanin implements the fan-in concurrency pattern: merging several
// input channels into a single output channel.
//
// Fan-in is how you recombine the results of work that was fanned out across
// many goroutines (a worker pool, several generators, multiple network calls).
// The classic implementation starts one forwarding goroutine per input and a
// sync.WaitGroup that closes the output exactly once, after every input has
// been fully drained.
package fanin

import (
	"context"
	"sync"
)

// Merge fans in any number of input channels into one output channel. Values
// from all inputs are interleaved in arrival order (the merge is not ordered).
// The output channel is closed once every input channel has been drained, or
// as soon as ctx is cancelled — whichever happens first.
//
// A forwarding goroutine is started per input; each exits when its input is
// closed or ctx is done, so no goroutine leaks even if the consumer stops
// reading early (provided it cancels ctx). Calling Merge with no channels
// returns an already-closed channel.
func Merge[T any](ctx context.Context, chans ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(chans))

	forward := func(in <-chan T) {
		defer wg.Done()
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return // input drained
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
	}

	for _, ch := range chans {
		go forward(ch)
	}

	// A single closer goroutine waits for all forwarders, then closes out
	// exactly once. Closing from any forwarder would risk a double close.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
