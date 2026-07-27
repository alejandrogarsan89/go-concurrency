package once_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/once"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGetComputesOnce(t *testing.T) {
	var calls int64
	lazy := once.New(func() int {
		atomic.AddInt64(&calls, 1)
		return 42
	})

	if got := lazy.Get(); got != 42 {
		t.Fatalf("Get() = %d, want 42", got)
	}
	if got := lazy.Get(); got != 42 {
		t.Fatalf("second Get() = %d, want 42", got)
	}
	if calls != 1 {
		t.Fatalf("init called %d times, want 1", calls)
	}
}

func TestGetConcurrentSingleInit(t *testing.T) {
	var calls int64
	lazy := once.New(func() int {
		atomic.AddInt64(&calls, 1)
		return 7
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([]int, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			results[i] = lazy.Get()
		}()
	}
	wg.Wait()

	if calls != 1 {
		t.Fatalf("init called %d times under concurrency, want 1", calls)
	}
	for i, v := range results {
		if v != 7 {
			t.Fatalf("results[%d] = %d, want 7", i, v)
		}
	}
}
