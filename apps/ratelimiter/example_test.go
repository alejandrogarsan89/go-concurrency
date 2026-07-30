package ratelimiter_test

import (
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/apps/ratelimiter"
)

func ExampleLimiter_Allow() {
	// 1 event/sec on average, but a burst of up to 2 back-to-back. A low rate
	// keeps this example deterministic: a refilled token is a full second away,
	// so the three calls below always observe the initial burst of two.
	l := ratelimiter.New(1, 2)

	// The bucket starts full, so the first two Allow calls succeed and the
	// third is refused until the bucket refills.
	fmt.Println(l.Allow())
	fmt.Println(l.Allow())
	fmt.Println(l.Allow())
	// Output:
	// true
	// true
	// false
}
