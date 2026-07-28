package psort_test

import (
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/psort"
)

func ExampleSort() {
	nums := []int{5, 2, 9, 1, 7, 3, 8, 4, 6, 0}
	psort.Sort(nums, func(a, b int) bool { return a < b })
	fmt.Println(nums)
	// Output:
	// [0 1 2 3 4 5 6 7 8 9]
}
