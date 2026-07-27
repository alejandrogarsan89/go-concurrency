package main

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/patterns/generator"
	"github.com/spf13/cobra"
)

func generatorCmd() *cobra.Command {
	var n, take int
	cmd := &cobra.Command{
		Use:   "generator",
		Short: "Stream values from a goroutine over a channel, optionally limited",
		Long: "Produces the integers 0..n-1 on a channel and prints them. With\n" +
			"--take k it only consumes the first k values and cancels the rest,\n" +
			"showing how a context stops the producer without leaking goroutines.\n\n" +
			"Example: demo generator --n 10 --take 3",
		RunE: func(_ *cobra.Command, _ []string) error {
			if n <= 0 {
				n = 10
				fmt.Println("No positive --n given, using sample: 10")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			stream := generator.Ints(ctx, n)
			if take > 0 {
				stream = generator.Take(ctx, stream, take)
				fmt.Printf("taking the first %d of %d values:\n", take, n)
			} else {
				fmt.Printf("streaming all %d values:\n", n)
			}
			for v := range stream {
				fmt.Print(v, " ")
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().IntVar(&n, "n", 10, "how many integers to generate")
	cmd.Flags().IntVar(&take, "take", 0, "only consume the first k values (0 = all)")
	return cmd
}
