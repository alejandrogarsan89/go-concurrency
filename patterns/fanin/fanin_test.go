package fanin_test

import (
	"context"
	"sort"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/fanin"
	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestMergeCombinesAllValues(t *testing.T) {
	ctx := context.Background()
	a := generator.FromSlice(ctx, []int{1, 2, 3})
	b := generator.FromSlice(ctx, []int{4, 5, 6})
	c := generator.FromSlice(ctx, []int{7, 8, 9})

	got := []int{}
	for v := range fanin.Merge(ctx, a, b, c) {
		got = append(got, v)
	}
	sort.Ints(got) // merge order is non-deterministic
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if len(got) != len(want) {
		t.Fatalf("merged %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("merged %v, want %v", got, want)
		}
	}
}

func TestMergeNoInputs(t *testing.T) {
	// Must return a closed channel, so the range loop ends immediately.
	count := 0
	for range fanin.Merge[int](context.Background()) {
		count++
	}
	if count != 0 {
		t.Fatalf("expected no values, got %d", count)
	}
}

func TestMergeSingleInput(t *testing.T) {
	ctx := context.Background()
	src := generator.Ints(ctx, 5)
	count := 0
	for range fanin.Merge(ctx, src) {
		count++
	}
	if count != 5 {
		t.Fatalf("merged %d values, want 5", count)
	}
}

func drain[T any](ch <-chan T) int {
	n := 0
	for range ch {
		n++
	}
	return n
}

func TestMergeCancellationDoesNotLeak(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Large upstream generators that we will abandon after a few reads.
	a := generator.Ints(ctx, 1_000_000)
	b := generator.Ints(ctx, 1_000_000)
	merged := fanin.Merge(ctx, a, b)

	<-merged
	<-merged
	cancel()
	// Drain until closed; all forwarders and generators must exit (goleak).
	drain(merged)
}
