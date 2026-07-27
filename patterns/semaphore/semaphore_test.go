package semaphore_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/semaphore"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestBoundsConcurrency(t *testing.T) {
	const limit = 3
	sem := semaphore.New(limit)
	ctx := context.Background()

	var inFlight, maxInFlight int64
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := sem.Acquire(ctx); err != nil {
				return
			}
			defer sem.Release()

			cur := atomic.AddInt64(&inFlight, 1)
			for {
				old := atomic.LoadInt64(&maxInFlight)
				if cur <= old || atomic.CompareAndSwapInt64(&maxInFlight, old, cur) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&inFlight, -1)
		}()
	}
	wg.Wait()

	if maxInFlight > limit {
		t.Fatalf("observed %d concurrent holders, want <= %d", maxInFlight, limit)
	}
}

func TestClampsToOne(t *testing.T) {
	sem := semaphore.New(0)
	if !sem.TryAcquire() {
		t.Fatal("first TryAcquire should succeed on a clamped semaphore")
	}
	if sem.TryAcquire() {
		t.Fatal("second TryAcquire should fail: only one slot")
	}
	sem.Release()
}

func TestTryAcquire(t *testing.T) {
	sem := semaphore.New(1)
	if !sem.TryAcquire() {
		t.Fatal("TryAcquire should succeed when a slot is free")
	}
	if sem.TryAcquire() {
		t.Fatal("TryAcquire should fail when no slot is free")
	}
	sem.Release()
	if !sem.TryAcquire() {
		t.Fatal("TryAcquire should succeed after Release")
	}
	sem.Release()
}

func TestAcquireRespectsContext(t *testing.T) {
	sem := semaphore.New(1)
	if err := sem.Acquire(context.Background()); err != nil {
		t.Fatalf("first Acquire: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := sem.Acquire(ctx) // no slot free -> should time out
	if err == nil {
		t.Fatal("Acquire should fail once the context is cancelled")
	}
	sem.Release()
}

func TestReleaseWithoutAcquirePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Release without a held slot should panic")
		}
	}()
	semaphore.New(1).Release()
}
