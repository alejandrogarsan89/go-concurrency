package pipeline_test

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/alejandrogarsan89/go-concurrency/patterns/pipeline"
)

func ExampleMap() {
	ctx := context.Background()
	src := generator.Ints(ctx, 5)
	for v := range pipeline.Map(ctx, src, func(v int) int { return v * v }) {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output:
	// 0 1 4 9 16
}

func Example_composed() {
	ctx := context.Background()
	src := generator.Ints(ctx, 10)
	// Stage 1: square. Stage 2: keep values divisible by 3.
	squared := pipeline.Map(ctx, src, func(v int) int { return v * v })
	out := pipeline.Filter(ctx, squared, func(v int) bool { return v%3 == 0 })
	for v := range out {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output:
	// 0 9 36 81
}
