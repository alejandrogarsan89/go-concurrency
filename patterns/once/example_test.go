package once_test

import (
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/once"
)

func ExampleLazy() {
	// The expensive computation runs only on the first Get, then is cached.
	config := once.New(func() map[string]string {
		fmt.Println("loading config...") // side effect proves it runs once
		return map[string]string{"env": "prod"}
	})

	fmt.Println(config.Get()["env"])
	fmt.Println(config.Get()["env"]) // no second "loading config..."
	// Output:
	// loading config...
	// prod
	// prod
}
