package main

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/mapreduce"
	"github.com/spf13/cobra"
)

func mapreduceCmd() *cobra.Command {
	var n, workers int
	cmd := &cobra.Command{
		Use:   "mapreduce",
		Short: "Parallel map-reduce: measure the speedup over a serial fold",
		Long: "Folds a CPU-bound function over --n items, first serially and then with\n" +
			"MapReduce across --workers goroutines, and prints the measured speedup.\n" +
			"On a multi-core machine the parallel run should be several times faster.\n\n" +
			"Example: demo mapreduce --n 200000 --workers 0   (0 = all CPUs)",
		RunE: func(_ *cobra.Command, _ []string) error {
			if n <= 0 {
				n = 200_000
			}
			if workers <= 0 {
				workers = runtime.GOMAXPROCS(0)
			}
			inputs := make([]int, n)
			for i := range inputs {
				inputs[i] = i
			}
			mapper := func(x int) float64 {
				acc := 0.0
				for i := 0; i < 200; i++ {
					acc += math.Sqrt(float64(x*i + 1))
				}
				return acc
			}
			add := func(a, b float64) float64 { return a + b }

			start := time.Now()
			serial := 0.0
			for _, x := range inputs {
				serial = add(serial, mapper(x))
			}
			serialDur := time.Since(start)

			start = time.Now()
			parallel := mapreduce.MapReduce(context.Background(), inputs, workers, mapper, add, 0.0)
			parallelDur := time.Since(start)

			fmt.Printf("items:    %d\n", n)
			fmt.Printf("workers:  %d (GOMAXPROCS=%d)\n", workers, runtime.GOMAXPROCS(0))
			fmt.Printf("serial:   %v\n", serialDur.Round(time.Microsecond))
			fmt.Printf("parallel: %v\n", parallelDur.Round(time.Microsecond))
			fmt.Printf("speedup:  %.2fx\n", float64(serialDur)/float64(parallelDur))
			// Floating-point addition is not associative, so the parallel fold's
			// different summation order yields a result that matches only to
			// within rounding error — itself a lesson in parallel reduction.
			if math.Abs(serial-parallel) > 1e-6*math.Abs(serial) {
				return fmt.Errorf("results differ beyond rounding: serial=%v parallel=%v", serial, parallel)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&n, "n", 200_000, "number of items to fold")
	cmd.Flags().IntVar(&workers, "workers", 0, "worker goroutines (0 = all CPUs)")
	return cmd
}
