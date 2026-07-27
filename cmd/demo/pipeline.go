package main

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/alejandrogarsan89/go-concurrency/patterns/pipeline"
	"github.com/spf13/cobra"
)

func pipelineCmd() *cobra.Command {
	var n int
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Multi-stage pipeline: generate -> square -> keep evens",
		Long: "Streams 0..n-1 through a two-stage pipeline (map then filter) where\n" +
			"each stage is its own goroutine connected by channels.\n\n" +
			"Example: demo pipeline --n 12",
		RunE: func(_ *cobra.Command, _ []string) error {
			if n <= 0 {
				n = 12
			}
			ctx := context.Background()

			src := generator.Ints(ctx, n)
			squared := pipeline.Map(ctx, src, func(v int) int { return v * v })
			evens := pipeline.Filter(ctx, squared, func(v int) bool { return v%2 == 0 })

			fmt.Printf("0..%d  --map(x^2)-->  --filter(even)-->\n", n-1)
			for v := range evens {
				fmt.Print(v, " ")
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().IntVar(&n, "n", 12, "how many integers to feed the pipeline")
	return cmd
}
