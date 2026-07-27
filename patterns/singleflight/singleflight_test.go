package singleflight_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/singleflight"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDoReturnsValue(t *testing.T) {
	var g singleflight.Group[string, int]
	got, err, shared := g.Do("k", func() (int, error) { return 99, nil })
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if got != 99 {
		t.Fatalf("Do() = %d, want 99", got)
	}
	if shared {
		t.Fatal("a lone call should not be shared")
	}
}

func TestDoPropagatesError(t *testing.T) {
	var g singleflight.Group[string, int]
	errBoom := errors.New("boom")
	_, err, _ := g.Do("k", func() (int, error) { return 0, errBoom })
	if !errors.Is(err, errBoom) {
		t.Fatalf("Do() error = %v, want %v", err, errBoom)
	}
}

func TestDoDeduplicatesConcurrentCalls(t *testing.T) {
	var g singleflight.Group[string, int]
	var calls int64

	// Block the owner until all callers have joined, guaranteeing overlap.
	release := make(chan struct{})
	const callers = 50
	var started sync.WaitGroup
	started.Add(callers)

	var wg sync.WaitGroup
	results := make([]int, callers)
	sharedFlags := make([]bool, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			started.Done()
			v, _, shared := g.Do("same-key", func() (int, error) {
				atomic.AddInt64(&calls, 1)
				<-release // hold the call open
				return 123, nil
			})
			results[i] = v
			sharedFlags[i] = shared
		}()
	}

	started.Wait() // all goroutines have entered Do
	close(release) // let the single owner finish
	wg.Wait()

	if calls != 1 {
		t.Fatalf("fn executed %d times, want exactly 1", calls)
	}
	sharedCount := 0
	for i := 0; i < callers; i++ {
		if results[i] != 123 {
			t.Fatalf("results[%d] = %d, want 123", i, results[i])
		}
		if sharedFlags[i] {
			sharedCount++
		}
	}
	if sharedCount == 0 {
		t.Fatal("expected at least one caller to report a shared result")
	}
}

func TestKeyFreedAfterCall(t *testing.T) {
	var g singleflight.Group[string, int]
	var calls int64
	fn := func() (int, error) {
		atomic.AddInt64(&calls, 1)
		return 1, nil
	}
	g.Do("k", fn)
	g.Do("k", fn) // sequential -> not deduplicated
	if calls != 2 {
		t.Fatalf("fn executed %d times, want 2 (key freed between calls)", calls)
	}
}

func TestDoPanicIsReRaisedNotLeaked(t *testing.T) {
	var g singleflight.Group[string, int]

	// Concurrent callers whose shared fn panics: every caller must recover the
	// panic rather than block forever, and the key must not be poisoned.
	release := make(chan struct{})
	const callers = 20
	var started sync.WaitGroup
	started.Add(callers)

	var wg sync.WaitGroup
	var panics int64
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt64(&panics, 1)
				}
			}()
			started.Done()
			g.Do("boom", func() (int, error) {
				<-release
				panic("kaboom")
			})
		}()
	}

	started.Wait()
	close(release)
	wg.Wait()

	if panics != callers {
		t.Fatalf("recovered %d panics, want %d (waiters must re-raise, not leak)", panics, callers)
	}

	// Key must be usable again after the panic (not poisoned).
	got, err, _ := g.Do("boom", func() (int, error) { return 7, nil })
	if err != nil || got != 7 {
		t.Fatalf("after panic, Do = (%d, %v), want (7, nil)", got, err)
	}
}
