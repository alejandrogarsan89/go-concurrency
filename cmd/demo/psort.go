package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"slices"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/psort"
	"github.com/spf13/cobra"
)

func psortCmd() *cobra.Command {
	var n int
	cmd := &cobra.Command{
		Use:   "psort",
		Short: "Parallel merge sort: measure the speedup over the stdlib sort",
		Long: "Sorts --n random integers with the single-threaded standard library\n" +
			"sort and with the parallel merge sort, and prints the speedup. Sorting\n" +
			"is partly memory-bandwidth bound, so the speedup is real but below the\n" +
			"core count — a nice illustration of Amdahl's law.\n\n" +
			"Example: demo psort --n 2000000",
		RunE: func(_ *cobra.Command, _ []string) error {
			if n <= 0 {
				n = 2_000_000
			}
			base := make([]int, n)
			r := rand.New(rand.NewSource(1))
			for i := range base {
				base[i] = r.Int()
			}

			stdData := append([]int(nil), base...)
			start := time.Now()
			slices.Sort(stdData)
			stdDur := time.Since(start)

			parData := append([]int(nil), base...)
			start = time.Now()
			psort.Sort(parData, func(a, b int) bool { return a < b })
			parDur := time.Since(start)

			fmt.Printf("items:    %d\n", n)
			fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
			fmt.Printf("stdlib:   %v\n", stdDur.Round(time.Microsecond))
			fmt.Printf("parallel: %v\n", parDur.Round(time.Microsecond))
			fmt.Printf("speedup:  %.2fx\n", float64(stdDur)/float64(parDur))
			if !slices.Equal(stdData, parData) {
				return fmt.Errorf("parallel sort produced a different result")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&n, "n", 2_000_000, "number of integers to sort")
	return cmd
}
