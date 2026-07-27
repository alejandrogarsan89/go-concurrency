package generator_test

import (
	"context"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
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

func TestInts(t *testing.T) {
	got := collect(generator.Ints(context.Background(), 5))
	want := []int{0, 1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("Ints produced %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Ints produced %v, want %v", got, want)
		}
	}
}

func TestIntsZeroOrNegative(t *testing.T) {
	if got := collect(generator.Ints(context.Background(), 0)); len(got) != 0 {
		t.Fatalf("Ints(0) = %v, want empty", got)
	}
	if got := collect(generator.Ints(context.Background(), -3)); len(got) != 0 {
		t.Fatalf("Ints(-3) = %v, want empty", got)
	}
}

func TestFromSlice(t *testing.T) {
	got := collect(generator.FromSlice(context.Background(), []string{"a", "b", "c"}))
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("FromSlice = %v", got)
	}
}

func TestTakeLimits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // stop the upstream generator once we stop reading it
	src := generator.Ints(ctx, 100)
	got := collect(generator.Take(ctx, src, 3))
	if len(got) != 3 || got[0] != 0 || got[2] != 2 {
		t.Fatalf("Take(3) = %v, want [0 1 2]", got)
	}
}

func TestTakeMoreThanAvailable(t *testing.T) {
	src := generator.FromSlice(context.Background(), []int{1, 2})
	got := collect(generator.Take(context.Background(), src, 10))
	if len(got) != 2 {
		t.Fatalf("Take beyond source = %v, want 2 values", got)
	}
}

func TestTakeZero(t *testing.T) {
	src := generator.Ints(context.Background(), 10)
	got := collect(generator.Take(context.Background(), src, 0))
	if len(got) != 0 {
		t.Fatalf("Take(0) = %v, want empty", got)
	}
	// Drain the source so its goroutine exits and goleak stays happy.
	collect(src)
}

func TestCancellationStopsGeneratorWithoutLeak(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := generator.Ints(ctx, 1_000_000)
	// Read a couple of values, then cancel and stop reading. The producer
	// goroutine must observe cancellation and exit (verified by goleak).
	<-ch
	<-ch
	cancel()
	// Drain whatever is already buffered/in-flight until closed.
	collect(ch)
}
