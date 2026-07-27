package generator_test

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
)

func ExampleInts() {
	for v := range generator.Ints(context.Background(), 4) {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output:
	// 0 1 2 3
}

func ExampleTake() {
	// Take a bounded prefix of a large generator without producing the rest.
	// Cancelling the context stops the upstream generator so it does not leak.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	src := generator.Ints(ctx, 1_000_000)
	for v := range generator.Take(ctx, src, 3) {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output:
	// 0 1 2
}
