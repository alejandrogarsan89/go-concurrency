package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/singleflight"
	"github.com/spf13/cobra"
)

func singleflightCmd() *cobra.Command {
	var callers int
	cmd := &cobra.Command{
		Use:   "singleflight",
		Short: "De-duplicate concurrent identical calls (thundering-herd protection)",
		Long: "Fires --callers goroutines that all request the same key at once. The\n" +
			"expensive function counts its executions: it runs a single time and every\n" +
			"caller shares that one result.\n\n" +
			"Example: demo singleflight --callers 100",
		RunE: func(_ *cobra.Command, _ []string) error {
			if callers <= 0 {
				callers = 100
			}
			var g singleflight.Group[string, int]
			var executions int64

			var wg sync.WaitGroup
			wg.Add(callers)
			for i := 0; i < callers; i++ {
				go func() {
					defer wg.Done()
					_, _, _ = g.Do("expensive-key", func() (int, error) {
						atomic.AddInt64(&executions, 1)
						time.Sleep(20 * time.Millisecond) // simulate expensive work
						return 42, nil
					})
				}()
			}
			wg.Wait()

			fmt.Printf("%d concurrent callers; expensive function executed %d time(s)\n", callers, executions)
			return nil
		},
	}
	cmd.Flags().IntVar(&callers, "callers", 100, "number of concurrent callers for the same key")
	return cmd
}
