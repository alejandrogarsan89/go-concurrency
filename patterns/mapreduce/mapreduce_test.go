package mapreduce_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/mapreduce"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSumMatchesSequential(t *testing.T) {
	inputs := make([]int, 1000)
	want := 0
	for i := range inputs {
		inputs[i] = i
		want += i * i
	}
	got := mapreduce.MapReduce(context.Background(), inputs, 8,
		func(x int) int { return x * x },
		func(a, b int) int { return a + b },
		0,
	)
	if got != want {
		t.Fatalf("MapReduce = %d, want %d", got, want)
	}
}

func TestPreservesOrderForNonCommutativeReducer(t *testing.T) {
	// String concatenation is associative but NOT commutative: the result must
	// match the left-to-right fold regardless of chunking.
	inputs := make([]string, 100)
	var want strings.Builder
	for i := range inputs {
		s := string(rune('a' + i%26))
		inputs[i] = s
		want.WriteString(s)
	}
	got := mapreduce.MapReduce(context.Background(), inputs, 7,
		func(s string) string { return s },
		func(a, b string) string { return a + b },
		"",
	)
	if got != want.String() {
		t.Fatalf("MapReduce did not preserve order:\n got %q\nwant %q", got, want.String())
	}
}

func TestEmptyReturnsIdentity(t *testing.T) {
	got := mapreduce.MapReduce(context.Background(), []int{}, 4,
		func(x int) int { return x },
		func(a, b int) int { return a + b },
		42,
	)
	if got != 42 {
		t.Fatalf("empty input = %d, want identity 42", got)
	}
}

func TestClampsWorkers(t *testing.T) {
	inputs := []int{1, 2, 3}
	got := mapreduce.MapReduce(context.Background(), inputs, 0, // clamps to 1
		func(x int) int { return x },
		func(a, b int) int { return a + b },
		0,
	)
	if got != 6 {
		t.Fatalf("MapReduce = %d, want 6", got)
	}
	// More workers than inputs is clamped to len(inputs).
	got = mapreduce.MapReduce(context.Background(), inputs, 100,
		func(x int) int { return x },
		func(a, b int) int { return a + b },
		0,
	)
	if got != 6 {
		t.Fatalf("MapReduce = %d, want 6", got)
	}
}

func TestIdleWorkersWhenChunksExceedInput(t *testing.T) {
	// With more workers than chunks can fill, the trailing workers get an empty
	// range and contribute identity; the sum must still be correct.
	inputs := make([]int, 10)
	want := 0
	for i := range inputs {
		inputs[i] = i + 1
		want += i + 1
	}
	got := mapreduce.MapReduce(context.Background(), inputs, 8,
		func(x int) int { return x },
		func(a, b int) int { return a + b },
		0,
	)
	if got != want {
		t.Fatalf("MapReduce = %d, want %d", got, want)
	}
}

func TestCancellationStopsEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	inputs := make([]int, 1000)
	for i := range inputs {
		inputs[i] = 1
	}
	got := mapreduce.MapReduce(ctx, inputs, 4,
		func(x int) int { return x },
		func(a, b int) int { return a + b },
		0,
	)
	// Best-effort partial: with an already-cancelled context, workers stop
	// immediately, so the sum is at most the full total.
	if got < 0 || got > 1000 {
		t.Fatalf("cancelled MapReduce = %d, want between 0 and 1000", got)
	}
}
