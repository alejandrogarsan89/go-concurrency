package waitgroup_test

import (
	"fmt"
	"sort"

	"github.com/alejandrogarsan89/go-concurrency/patterns/waitgroup"
)

func ExampleRunAll() {
	done := make(chan string, 2)
	waitgroup.RunAll(
		func() { done <- "task-a" },
		func() { done <- "task-b" },
	)
	close(done)
	got := []string{}
	for s := range done {
		got = append(got, s)
	}
	sort.Strings(got) // completion order is non-deterministic
	fmt.Println(got)
	// Output:
	// [task-a task-b]
}

func ExampleMap() {
	squares := waitgroup.Map([]int{1, 2, 3, 4}, func(v int) int { return v * v })
	fmt.Println(squares) // order matches the input
	// Output:
	// [1 4 9 16]
}
