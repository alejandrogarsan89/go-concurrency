package main

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/fanin"
	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/spf13/cobra"
)

func faninCmd() *cobra.Command {
	var sources, per int
	cmd := &cobra.Command{
		Use:   "fanin",
		Short: "Merge several channels into one (fan-in)",
		Long: "Starts --sources generators, each emitting --per values, then merges\n" +
			"them into a single stream and prints the interleaved result.\n\n" +
			"Example: demo fanin --sources 3 --per 4",
		RunE: func(_ *cobra.Command, _ []string) error {
			if sources <= 0 {
				sources = 3
			}
			if per <= 0 {
				per = 4
			}
			ctx := context.Background()
			chans := make([]<-chan int, sources)
			for s := 0; s < sources; s++ {
				base := s * 100
				vals := make([]int, per)
				for i := range vals {
					vals[i] = base + i
				}
				chans[s] = generator.FromSlice(ctx, vals)
			}

			count := 0
			fmt.Printf("merging %d sources x %d values:\n", sources, per)
			for v := range fanin.Merge(ctx, chans...) {
				fmt.Print(v, " ")
				count++
			}
			fmt.Printf("\nreceived %d values total\n", count)
			return nil
		},
	}
	cmd.Flags().IntVar(&sources, "sources", 3, "number of source channels")
	cmd.Flags().IntVar(&per, "per", 4, "values emitted per source")
	return cmd
}
