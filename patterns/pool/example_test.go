package pool_test

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/alejandrogarsan89/go-concurrency/patterns/pool"
)

func ExampleMap() {
	// Bounded parallel map: at most 3 workers, results in input order.
	words := []string{"go", "concurrency", "worker", "pool"}
	lengths := pool.Map(context.Background(), words, 3, func(_ context.Context, w string) int {
		return len(w)
	})
	fmt.Println(lengths)
	// Output:
	// [2 11 6 4]
}

func ExampleProcess() {
	ctx := context.Background()
	in := generator.FromSlice(ctx, []string{"a", "b", "c"})
	out := pool.Process(ctx, in, 2, func(_ context.Context, s string) string {
		return strings.ToUpper(s)
	})

	got := []string{}
	for v := range out {
		got = append(got, v)
	}
	sort.Strings(got) // Process results are unordered
	fmt.Println(got)
	// Output:
	// [A B C]
}
