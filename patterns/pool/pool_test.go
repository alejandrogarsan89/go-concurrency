package pool_test

import (
	"context"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/alejandrogarsan89/go-concurrency/patterns/pool"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func drain[T any](ch <-chan T) int {
	n := 0
	for range ch {
		n++
	}
	return n
}

func TestProcessHandlesEveryValue(t *testing.T) {
	ctx := context.Background()
	in := generator.Ints(ctx, 100)
	out := pool.Process(ctx, in, 8, func(_ context.Context, v int) int { return v * 2 })

	got := []int{}
	for v := range out {
		got = append(got, v)
	}
	if len(got) != 100 {
		t.Fatalf("got %d results, want 100", len(got))
	}
	sort.Ints(got)
	for i, v := range got {
		if v != i*2 {
			t.Fatalf("got[%d] = %d, want %d", i, v, i*2)
		}
	}
}

func TestProcessClampsWorkers(t *testing.T) {
	ctx := context.Background()
	in := generator.Ints(ctx, 10)
	out := pool.Process(ctx, in, 0, func(_ context.Context, v int) int { return v })
	count := 0
	for range out {
		count++
	}
	if count != 10 {
		t.Fatalf("got %d, want 10 (workers should clamp to 1)", count)
	}
}

func TestProcessBoundsConcurrency(t *testing.T) {
	ctx := context.Background()
	const workers = 4
	var inFlight, maxInFlight int64

	in := generator.Ints(ctx, 200)
	out := pool.Process(ctx, in, workers, func(_ context.Context, v int) int {
		cur := atomic.AddInt64(&inFlight, 1)
		for {
			old := atomic.LoadInt64(&maxInFlight)
			if cur <= old || atomic.CompareAndSwapInt64(&maxInFlight, old, cur) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		atomic.AddInt64(&inFlight, -1)
		return v
	})
	drain(out)
	if maxInFlight > workers {
		t.Fatalf("observed %d concurrent workers, want <= %d", maxInFlight, workers)
	}
}

func TestProcessCancellationNoLeak(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := generator.Ints(ctx, 1_000_000)
	out := pool.Process(ctx, in, 4, func(_ context.Context, v int) int { return v })
	<-out
	<-out
	cancel()
	drain(out)
}

func TestMapPreservesOrder(t *testing.T) {
	inputs := make([]int, 50)
	for i := range inputs {
		inputs[i] = i
	}
	got := pool.Map(context.Background(), inputs, 8, func(_ context.Context, v int) int {
		return v * v
	})
	for i := range inputs {
		if got[i] != i*i {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], i*i)
		}
	}
}

func TestMapEmpty(t *testing.T) {
	got := pool.Map(context.Background(), []int{}, 4, func(_ context.Context, v int) int { return v })
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

func TestMapClampsWorkers(t *testing.T) {
	got := pool.Map(context.Background(), []int{1, 2, 3}, -1, func(_ context.Context, v int) int {
		return v + 1
	})
	want := []int{2, 3, 4}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Map = %v, want %v", got, want)
		}
	}
}
