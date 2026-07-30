// Package ratelimiter implements a token-bucket rate limiter: work is allowed to
// proceed at a steady average rate, with a configurable burst.
//
// A token bucket fills at rate tokens per second up to a maximum (the burst
// size). Each unit of work consumes one token; if none is available, Allow
// reports false and Wait blocks (cancellably) until the bucket refills. Unlike a
// semaphore — which bounds how many things run at *once* — a rate limiter bounds
// how *often* they may start, which is what most external APIs actually enforce.
//
// The bucket level is computed lazily from elapsed time, so there is no
// background goroutine to leak. The clock is injectable, making the limiter fully
// deterministic under test.
package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter. It is safe for concurrent use.
type Limiter struct {
	mu     sync.Mutex
	rate   float64 // tokens added per second
	burst  float64 // maximum tokens the bucket can hold
	tokens float64 // current tokens
	last   time.Time
	now    func() time.Time
}

// New returns a Limiter that permits ratePerSec events per second on average,
// allowing short bursts of up to burst events. The bucket starts full. ratePerSec
// is clamped to a tiny positive value and burst to at least 1.
func New(ratePerSec float64, burst int) *Limiter {
	return newWithClock(ratePerSec, burst, time.Now)
}

func newWithClock(ratePerSec float64, burst int, now func() time.Time) *Limiter {
	if ratePerSec <= 0 {
		ratePerSec = 1e-9
	}
	if burst < 1 {
		burst = 1
	}
	b := float64(burst)
	return &Limiter{
		rate:   ratePerSec,
		burst:  b,
		tokens: b,
		last:   now(),
		now:    now,
	}
}

// refill adds the tokens accrued since the last update. The caller must hold mu.
func (l *Limiter) refill() {
	now := l.now()
	elapsed := now.Sub(l.last).Seconds()
	if elapsed > 0 {
		l.tokens += elapsed * l.rate
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
		l.last = now
	}
}

// Allow reports whether one event may proceed right now, consuming a token if so.
// It never blocks.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available and consumes it, or until ctx is
// cancelled (in which case it returns ctx.Err() and consumes nothing). It is the
// blocking counterpart to Allow.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		l.mu.Lock()
		l.refill()
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Time until the next whole token becomes available.
		deficit := 1 - l.tokens
		wait := time.Duration(deficit / l.rate * float64(time.Second))
		l.mu.Unlock()

		if wait <= 0 {
			wait = time.Millisecond
		}
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			// loop and try again
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}
