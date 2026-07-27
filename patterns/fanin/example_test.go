package fanin_test

import (
	"context"
	"fmt"
	"sort"

	"github.com/alejandrogarsan89/go-concurrency/patterns/fanin"
	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
)

func ExampleMerge() {
	ctx := context.Background()
	evens := generator.FromSlice(ctx, []int{2, 4, 6})
	odds := generator.FromSlice(ctx, []int{1, 3, 5})

	got := []int{}
	for v := range fanin.Merge(ctx, evens, odds) {
		got = append(got, v)
	}
	sort.Ints(got) // interleaving order is not deterministic
	fmt.Println(got)
	// Output:
	// [1 2 3 4 5 6]
}
