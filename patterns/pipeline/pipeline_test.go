package pipeline_test

import (
	"context"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/alejandrogarsan89/go-concurrency/patterns/pipeline"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func collect[T any](ch <-chan T) []T {
	out := []T{}
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func TestMapTransformsInOrder(t *testing.T) {
	ctx := context.Background()
	src := generator.Ints(ctx, 5)
	got := collect(pipeline.Map(ctx, src, func(v int) int { return v * 10 }))
	want := []int{0, 10, 20, 30, 40}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Map = %v, want %v", got, want)
		}
	}
}

func TestMapChangesType(t *testing.T) {
	ctx := context.Background()
	src := generator.FromSlice(ctx, []int{1, 2, 3})
	got := collect(pipeline.Map(ctx, src, func(v int) bool { return v%2 == 0 }))
	want := []bool{false, true, false}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Map = %v, want %v", got, want)
		}
	}
}

func TestFilterKeepsMatching(t *testing.T) {
	ctx := context.Background()
	src := generator.Ints(ctx, 10)
	got := collect(pipeline.Filter(ctx, src, func(v int) bool { return v%2 == 0 }))
	want := []int{0, 2, 4, 6, 8}
	if len(got) != len(want) {
		t.Fatalf("Filter = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Filter = %v, want %v", got, want)
		}
	}
}

func TestComposedStagesPreserveOrder(t *testing.T) {
	ctx := context.Background()
	src := generator.Ints(ctx, 10)
	// square -> keep even -> should yield squares of even inputs, in order.
	squared := pipeline.Map(ctx, src, func(v int) int { return v * v })
	evens := pipeline.Filter(ctx, squared, func(v int) bool { return v%2 == 0 })
	got := collect(evens)
	want := []int{0, 4, 16, 36, 64}
	if len(got) != len(want) {
		t.Fatalf("composed = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("composed = %v, want %v", got, want)
		}
	}
}

func TestCancellationNoLeak(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := generator.Ints(ctx, 1_000_000)
	stage := pipeline.Map(ctx, src, func(v int) int { return v + 1 })
	<-stage
	<-stage
	cancel()
	collect(stage) // drain until closed so all stage goroutines exit
}

func TestFilterCancellationNoLeak(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := generator.Ints(ctx, 1_000_000)
	stage := pipeline.Filter(ctx, src, func(v int) bool { return v%2 == 0 })
	<-stage
	<-stage
	cancel()
	collect(stage)
}
