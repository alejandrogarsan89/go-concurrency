package ratelimiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// fakeClock is a manually advanced clock for deterministic tests.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func TestAllowConsumesBurstThenRefills(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	l := newWithClock(10, 3, clk.now) // 10 tokens/sec, burst 3, starts full

	// Burst of 3 is allowed immediately.
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("Allow() #%d = false, want true (within burst)", i+1)
		}
	}
	// Bucket now empty.
	if l.Allow() {
		t.Fatal("Allow() = true, want false (burst exhausted)")
	}
	// After 100ms at 10/sec, exactly one token accrues.
	clk.advance(100 * time.Millisecond)
	if !l.Allow() {
		t.Fatal("Allow() = false after refill, want true")
	}
	if l.Allow() {
		t.Fatal("Allow() = true, want false (only one token refilled)")
	}
}

func TestRefillIsCappedAtBurst(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	l := newWithClock(5, 2, clk.now) // burst 2
	// Drain the bucket.
	l.Allow()
	l.Allow()
	// Advance a long time; tokens must cap at burst, not accumulate unbounded.
	clk.advance(time.Hour)
	allowed := 0
	for i := 0; i < 10; i++ {
		if l.Allow() {
			allowed++
		}
	}
	if allowed != 2 {
		t.Fatalf("after long idle, allowed %d in a row, want burst=2", allowed)
	}
}

func TestWaitReturnsWhenTokenAvailable(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	l := newWithClock(100, 1, clk.now) // starts with 1 token
	if err := l.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() = %v, want nil (token available)", err)
	}
}

func TestWaitRespectsContext(t *testing.T) {
	// Real clock, empty bucket, very slow refill: Wait must observe cancellation.
	l := New(0.001, 1) // ~1 token per 1000s
	l.tokens = 0       // drain

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.Wait(ctx); err == nil {
		t.Fatal("Wait() = nil, want context error")
	}
}
