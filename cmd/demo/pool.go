package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/alejandrogarsan89/go-concurrency/patterns/pool"
	"github.com/spf13/cobra"
)

func poolCmd() *cobra.Command {
	var jobs, workers int
	cmd := &cobra.Command{
		Use:   "pool",
		Short: "Worker pool: process N jobs with a fixed number of workers",
		Long: "Feeds --jobs items through a pool of --workers goroutines. Each job\n" +
			"simulates work, so you can see that raising --workers finishes the\n" +
			"batch faster, up to the number of CPU cores.\n\n" +
			"Example: demo pool --jobs 20 --workers 4",
		RunE: func(_ *cobra.Command, _ []string) error {
			if jobs <= 0 {
				jobs = 20
			}
			if workers <= 0 {
				workers = 4
			}
			ctx := context.Background()
			in := generator.Ints(ctx, jobs)

			start := time.Now()
			out := pool.Process(ctx, in, workers, func(_ context.Context, v int) int {
				time.Sleep(20 * time.Millisecond) // simulate work
				return v * v
			})
			count := 0
			for range out {
				count++
			}
			elapsed := time.Since(start).Round(time.Millisecond)
			fmt.Printf("processed %d jobs with %d workers in %v\n", count, workers, elapsed)
			fmt.Printf("(serial would take ~%v)\n", time.Duration(jobs)*20*time.Millisecond)
			return nil
		},
	}
	cmd.Flags().IntVar(&jobs, "jobs", 20, "number of jobs to process")
	cmd.Flags().IntVar(&workers, "workers", 4, "number of concurrent workers")
	return cmd
}
