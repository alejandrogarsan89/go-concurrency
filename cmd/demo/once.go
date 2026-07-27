package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/alejandrogarsan89/go-concurrency/patterns/once"
	"github.com/spf13/cobra"
)

func onceCmd() *cobra.Command {
	var goroutines int
	cmd := &cobra.Command{
		Use:   "once",
		Short: "Lazy init: an expensive value is computed once, shared by many goroutines",
		Long: "Launches --goroutines callers that all ask a once.Lazy for its value at\n" +
			"the same time. The initializer counts how many times it actually runs —\n" +
			"always exactly one, no matter how many callers race.\n\n" +
			"Example: demo once --goroutines 100",
		RunE: func(_ *cobra.Command, _ []string) error {
			if goroutines <= 0 {
				goroutines = 100
			}
			var inits int64
			lazy := once.New(func() int {
				atomic.AddInt64(&inits, 1)
				return 42
			})

			var wg sync.WaitGroup
			wg.Add(goroutines)
			for i := 0; i < goroutines; i++ {
				go func() {
					defer wg.Done()
					_ = lazy.Get()
				}()
			}
			wg.Wait()

			fmt.Printf("%d goroutines called Get(); initializer ran %d time(s)\n", goroutines, inits)
			return nil
		},
	}
	cmd.Flags().IntVar(&goroutines, "goroutines", 100, "number of concurrent callers")
	return cmd
}
