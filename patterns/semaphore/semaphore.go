// Package semaphore implements a counting semaphore that bounds how many
// goroutines may hold a resource at once.
//
// It is built on the memory-model fact that a buffered channel of capacity N is
// a counting semaphore: sending a token blocks once N are outstanding, and
// receiving one releases a slot. This is the simplest way to cap concurrency
// around a shared resource — a connection pool, a rate-limited API, a bounded
// amount of memory — independently of how many goroutines you launch.
package semaphore

import "context"

// Semaphore is a counting semaphore with a fixed number of slots. The zero value
// is not usable; construct one with New. It is safe for concurrent use.
type Semaphore struct {
	tokens chan struct{}
}

// New returns a Semaphore with n concurrent slots. n is clamped to at least 1.
func New(n int) *Semaphore {
	if n < 1 {
		n = 1
	}
	return &Semaphore{tokens: make(chan struct{}, n)}
}

// Acquire takes a slot, blocking until one is free or ctx is cancelled. It
// returns ctx.Err() if the context is done before a slot is obtained, in which
// case no slot is held and Release must not be called for this acquisition.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.tokens <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire takes a slot without blocking. It returns true if a slot was
// acquired (in which case the caller must Release it) and false if all slots are
// currently in use.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.tokens <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release returns a previously acquired slot. Calling Release without a matching
// successful Acquire/TryAcquire panics, mirroring an unlock of an unlocked mutex
// and surfacing the bug rather than silently miscounting.
func (s *Semaphore) Release() {
	select {
	case <-s.tokens:
	default:
		panic("semaphore: Release called without a held slot")
	}
}
