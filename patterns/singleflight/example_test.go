package singleflight_test

import (
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/singleflight"
)

func ExampleGroup() {
	var g singleflight.Group[string, string]

	// Only one execution runs per key at a time; here the calls are sequential
	// so each runs, but concurrent duplicates would share a single result.
	val, err, _ := g.Do("user:42", func() (string, error) {
		return "Ada Lovelace", nil
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(val)
	// Output:
	// Ada Lovelace
}
