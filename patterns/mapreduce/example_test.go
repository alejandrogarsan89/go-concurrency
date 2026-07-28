package mapreduce_test

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/mapreduce"
)

func ExampleMapReduce() {
	// Sum of squares of 1..10, mapped and reduced in parallel across 4 workers.
	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := mapreduce.MapReduce(context.Background(), inputs, 4,
		func(x int) int { return x * x },    // map
		func(a, b int) int { return a + b }, // reduce
		0,                                   // identity
	)
	fmt.Println(sum)
	// Output:
	// 385
}
