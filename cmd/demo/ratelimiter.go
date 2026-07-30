package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/apps/ratelimiter"
	"github.com/spf13/cobra"
)

func ratelimiterCmd() *cobra.Command {
	var requests, burst int
	var rate float64
	cmd := &cobra.Command{
		Use:   "ratelimiter",
		Short: "Token-bucket rate limiter throttling a burst of requests",
		Long: "Fires --requests requests as fast as possible through a token bucket that\n" +
			"permits --rate per second with a burst of --burst, printing which are\n" +
			"allowed immediately and how the limiter paces the rest via Wait.\n\n" +
			"Example: demo ratelimiter --requests 10 --rate 5 --burst 3",
		RunE: func(_ *cobra.Command, _ []string) error {
			if requests < 1 {
				requests = 10
			}
			if rate <= 0 {
				rate = 5
			}
			if burst < 1 {
				burst = 3
			}
			l := ratelimiter.New(rate, burst)

			fmt.Printf("rate: %.1f/s  burst: %d  requests: %d\n\n", rate, burst, requests)

			// Phase 1: non-blocking Allow shows the burst then refusals.
			allowed := 0
			for i := 1; i <= requests; i++ {
				if l.Allow() {
					allowed++
					fmt.Printf("Allow  #%2d -> accepted\n", i)
				} else {
					fmt.Printf("Allow  #%2d -> throttled\n", i)
				}
			}
			fmt.Printf("\naccepted immediately: %d/%d (burst=%d)\n\n", allowed, requests, burst)

			// Phase 2: blocking Wait paces requests to the sustained rate.
			l2 := ratelimiter.New(rate, burst)
			start := time.Now()
			for i := 1; i <= requests; i++ {
				if err := l2.Wait(context.Background()); err != nil {
					return err
				}
				fmt.Printf("Wait   #%2d -> sent at %v\n", i, time.Since(start).Round(10*time.Millisecond))
			}
			fmt.Printf("\ntotal time for %d paced requests: %v\n", requests, time.Since(start).Round(10*time.Millisecond))
			return nil
		},
	}
	cmd.Flags().IntVar(&requests, "requests", 10, "number of requests to send")
	cmd.Flags().Float64Var(&rate, "rate", 5, "sustained rate (tokens per second)")
	cmd.Flags().IntVar(&burst, "burst", 3, "burst size (bucket capacity)")
	return cmd
}
