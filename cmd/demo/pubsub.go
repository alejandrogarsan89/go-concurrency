package main

import (
	"fmt"
	"sync"

	"github.com/alejandrogarsan89/go-concurrency/apps/pubsub"
	"github.com/spf13/cobra"
)

func pubsubCmd() *cobra.Command {
	var subscribers, messages int
	cmd := &cobra.Command{
		Use:   "pubsub",
		Short: "In-memory publish/subscribe broker fanning out to subscribers",
		Long: "Starts --subscribers subscribers on one topic, publishes --messages\n" +
			"messages, and shows that each subscriber receives its own copy of every\n" +
			"message — a fan-out with correct channel-close discipline.\n\n" +
			"Example: demo pubsub --subscribers 3 --messages 4",
		RunE: func(_ *cobra.Command, _ []string) error {
			if subscribers < 1 {
				subscribers = 3
			}
			if messages < 1 {
				messages = 4
			}
			b := pubsub.New[string]()
			defer b.Close()

			var wg sync.WaitGroup
			var mu sync.Mutex
			received := make(map[int]int)

			for id := 1; id <= subscribers; id++ {
				ch, _ := b.Subscribe("events", messages)
				wg.Add(1)
				go func(id int, ch <-chan string) {
					defer wg.Done()
					for msg := range ch {
						mu.Lock()
						received[id]++
						fmt.Printf("subscriber %d received %q\n", id, msg)
						mu.Unlock()
					}
				}(id, ch)
			}

			for i := 1; i <= messages; i++ {
				msg := fmt.Sprintf("event-%d", i)
				n := b.Publish("events", msg)
				fmt.Printf("published %q -> delivered to %d subscribers\n", msg, n)
			}

			b.Close() // closes all subscriber channels so the goroutines exit
			wg.Wait()

			fmt.Println("\nsummary:")
			for id := 1; id <= subscribers; id++ {
				fmt.Printf("  subscriber %d: %d/%d messages\n", id, received[id], messages)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&subscribers, "subscribers", 3, "number of subscribers")
	cmd.Flags().IntVar(&messages, "messages", 4, "number of messages to publish")
	return cmd
}
