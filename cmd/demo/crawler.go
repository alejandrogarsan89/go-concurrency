package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/apps/crawler"
	"github.com/spf13/cobra"
)

// latencyFetcher wraps a MapFetcher with an artificial per-page delay so the
// demo shows real concurrency: with N pages each taking d, a serial crawl needs
// ~N*d while a bounded-concurrent crawl finishes much faster.
type latencyFetcher struct {
	inner crawler.MapFetcher
	delay time.Duration
}

func (l latencyFetcher) Fetch(ctx context.Context, url string) ([]string, error) {
	select {
	case <-time.After(l.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return l.inner.Fetch(ctx, url)
}

func crawlerCmd() *cobra.Command {
	var workers int
	var delayMS int
	cmd := &cobra.Command{
		Use:   "crawler",
		Short: "Concurrent web crawler over an in-memory link graph",
		Long: "Crawls a small, offline link graph with bounded concurrency: at most\n" +
			"--workers fetches run at once, each URL is visited exactly once, and the\n" +
			"total time reflects the parallel speedup over a serial crawl.\n\n" +
			"Example: demo crawler --workers 4 --delay 50",
		RunE: func(_ *cobra.Command, _ []string) error {
			if workers < 1 {
				workers = 4
			}
			if delayMS < 0 {
				delayMS = 50
			}
			graph := crawler.MapFetcher{
				"/":    {"/a", "/b", "/c"},
				"/a":   {"/a/1", "/a/2"},
				"/b":   {"/b/1"},
				"/c":   {"/a", "/b"}, // shared links exercise de-duplication
				"/a/1": {},
				"/a/2": {"/a/1"}, // cycle-ish: already visited
				"/b/1": {},
			}
			f := latencyFetcher{inner: graph, delay: time.Duration(delayMS) * time.Millisecond}

			start := time.Now()
			results := crawler.Crawl(context.Background(), "/", 5, workers, f)
			elapsed := time.Since(start)

			fmt.Printf("workers:  %d\n", workers)
			fmt.Printf("delay:    %dms/page\n", delayMS)
			fmt.Printf("visited:  %d pages\n", len(results))
			fmt.Printf("elapsed:  %v\n", elapsed.Round(time.Millisecond))
			serial := time.Duration(len(results)*delayMS) * time.Millisecond
			fmt.Printf("serial:   ~%v (estimated)\n", serial)
			fmt.Println("pages:")
			urls := make([]string, 0, len(results))
			for _, r := range results {
				urls = append(urls, fmt.Sprintf("  depth %d: %s", r.Depth, r.URL))
			}
			sort.Strings(urls)
			for _, line := range urls {
				fmt.Println(line)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&workers, "workers", 4, "maximum concurrent fetches")
	cmd.Flags().IntVar(&delayMS, "delay", 50, "artificial fetch latency per page (ms)")
	return cmd
}
