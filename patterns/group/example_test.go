package group_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/group"
)

func ExampleGroup() {
	// Run independent tasks concurrently and wait for the first error.
	var g group.Group
	g.Go(func() error { return nil })
	g.Go(func() error { return nil })

	if err := g.Wait(); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("all tasks succeeded")
	// Output:
	// all tasks succeeded
}

func ExampleWithContext() {
	// The first failure cancels the shared context, so siblings stop early.
	g, ctx := group.WithContext(context.Background())
	g.Go(func() error { return errors.New("task 1 failed") })
	g.Go(func() error {
		<-ctx.Done() // observe the cancellation triggered by task 1
		return ctx.Err()
	})

	fmt.Println(g.Wait())
	// Output:
	// task 1 failed
}
