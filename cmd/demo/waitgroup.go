package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/waitgroup"
	"github.com/spf13/cobra"
)

func waitgroupCmd() *cobra.Command {
	var tasks int
	cmd := &cobra.Command{
		Use:   "waitgroup",
		Short: "Run N tasks concurrently and wait for all of them (sync.WaitGroup)",
		Long: "Launches --tasks goroutines that each simulate a little work, then\n" +
			"blocks until every one has finished. Also shows Map, which runs a\n" +
			"function over inputs in parallel and returns results in order.\n\n" +
			"Example: demo waitgroup --tasks 8",
		RunE: func(_ *cobra.Command, _ []string) error {
			if tasks <= 0 {
				tasks = 8
				fmt.Println("No positive --tasks given, using sample: 8")
			}
			results := make([]string, tasks)
			funcs := make([]func(), tasks)
			for i := 0; i < tasks; i++ {
				funcs[i] = func() {
					time.Sleep(10 * time.Millisecond) // simulate work
					results[i] = fmt.Sprintf("task-%d done", i)
				}
			}
			start := time.Now()
			waitgroup.RunAll(funcs...)
			fmt.Printf("all %d tasks finished in %v (concurrently)\n", tasks, time.Since(start).Round(time.Millisecond))

			nums := make([]int, tasks)
			for i := range nums {
				nums[i] = i + 1
			}
			squares := waitgroup.Map(nums, func(v int) int { return v * v })
			fmt.Printf("Map squares %v -> %v (order preserved)\n", nums, squares)

			sort.Strings(results)
			fmt.Println("sample result:", results[0])
			return nil
		},
	}
	cmd.Flags().IntVar(&tasks, "tasks", 8, "number of concurrent tasks")
	return cmd
}
