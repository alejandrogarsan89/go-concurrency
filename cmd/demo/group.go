package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/group"
	"github.com/spf13/cobra"
)

func groupCmd() *cobra.Command {
	var tasks int
	var fail bool
	cmd := &cobra.Command{
		Use:   "group",
		Short: "errgroup-style tasks: wait for all, cancel the rest on first error",
		Long: "Runs --tasks concurrent tasks. With --fail, one task returns an error;\n" +
			"the shared context is cancelled so the still-running siblings stop early,\n" +
			"and Wait returns that first error.\n\n" +
			"Example: demo group --tasks 5 --fail",
		RunE: func(_ *cobra.Command, _ []string) error {
			if tasks <= 0 {
				tasks = 5
			}
			g, ctx := group.WithContext(context.Background())

			var cancelled int
			for i := 0; i < tasks; i++ {
				i := i
				g.Go(func() error {
					if fail && i == 0 {
						return fmt.Errorf("task %d failed", i)
					}
					select {
					case <-ctx.Done():
						cancelled++
						return ctx.Err()
					case <-time.After(50 * time.Millisecond):
						return nil
					}
				})
			}

			err := g.Wait()
			if err != nil && !errors.Is(err, context.Canceled) {
				fmt.Printf("group finished with first error: %v\n", err)
			} else {
				fmt.Printf("group finished; all %d tasks succeeded\n", tasks)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&tasks, "tasks", 5, "number of concurrent tasks")
	cmd.Flags().BoolVar(&fail, "fail", false, "make one task fail to trigger cancellation")
	return cmd
}
