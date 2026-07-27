package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/semaphore"
	"github.com/spf13/cobra"
)

func semaphoreCmd() *cobra.Command {
	var tasks, limit int
	cmd := &cobra.Command{
		Use:   "semaphore",
		Short: "Counting semaphore: cap how many tasks run at once",
		Long: "Runs --tasks goroutines but lets only --limit of them hold the resource\n" +
			"at the same time. Reports the peak observed concurrency, which never\n" +
			"exceeds --limit.\n\n" +
			"Example: demo semaphore --tasks 20 --limit 3",
		RunE: func(_ *cobra.Command, _ []string) error {
			if tasks <= 0 {
				tasks = 20
			}
			if limit <= 0 {
				limit = 3
			}
			sem := semaphore.New(limit)
			ctx := context.Background()

			var inFlight, peak int64
			var wg sync.WaitGroup
			wg.Add(tasks)
			for i := 0; i < tasks; i++ {
				go func() {
					defer wg.Done()
					if err := sem.Acquire(ctx); err != nil {
						return
					}
					defer sem.Release()
					cur := atomic.AddInt64(&inFlight, 1)
					for {
						old := atomic.LoadInt64(&peak)
						if cur <= old || atomic.CompareAndSwapInt64(&peak, old, cur) {
							break
						}
					}
					time.Sleep(20 * time.Millisecond)
					atomic.AddInt64(&inFlight, -1)
				}()
			}
			wg.Wait()

			fmt.Printf("ran %d tasks with a limit of %d; peak concurrency was %d\n", tasks, limit, peak)
			return nil
		},
	}
	cmd.Flags().IntVar(&tasks, "tasks", 20, "number of tasks to run")
	cmd.Flags().IntVar(&limit, "limit", 3, "maximum concurrent tasks")
	return cmd
}
