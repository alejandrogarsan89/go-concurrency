package waitgroup_test

import (
	"sort"
	"sync/atomic"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/waitgroup"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRunAllRunsEveryTask(t *testing.T) {
	var counter int64
	tasks := make([]func(), 100)
	for i := range tasks {
		tasks[i] = func() { atomic.AddInt64(&counter, 1) }
	}
	waitgroup.RunAll(tasks...)
	if got := atomic.LoadInt64(&counter); got != 100 {
		t.Fatalf("ran %d tasks, want 100", got)
	}
}

func TestRunAllSkipsNilTasks(t *testing.T) {
	var counter int64
	waitgroup.RunAll(
		func() { atomic.AddInt64(&counter, 1) },
		nil,
		func() { atomic.AddInt64(&counter, 1) },
		nil,
	)
	if got := atomic.LoadInt64(&counter); got != 2 {
		t.Fatalf("counter = %d, want 2", got)
	}
}

func TestRunAllEmpty(_ *testing.T) {
	waitgroup.RunAll() // must return promptly without blocking
}

func TestMapPreservesOrder(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := waitgroup.Map(in, func(v int) int { return v * v })
	want := []int{1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Map result = %v, want %v", got, want)
		}
	}
}

func TestMapEmpty(t *testing.T) {
	got := waitgroup.Map([]string{}, func(s string) int { return len(s) })
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

func TestMapConcurrencyIsRaceFree(t *testing.T) {
	// Large input so the race detector exercises concurrent writes to the
	// results slice (distinct indices => must be race-free).
	in := make([]int, 1000)
	for i := range in {
		in[i] = i
	}
	got := waitgroup.Map(in, func(v int) int { return v + 1 })
	sort.Ints(got)
	for i, v := range got {
		if v != i+1 {
			t.Fatalf("got[%d] = %d, want %d", i, v, i+1)
		}
	}
}
