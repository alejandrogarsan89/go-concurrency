// Package crawler implements a concurrent web crawler that combines several of
// the library's patterns into a real application: bounded parallelism (a
// semaphore caps concurrent fetches), de-duplication of already-seen URLs (a
// mutex-guarded visited set), a depth limit, and cooperative cancellation via
// context.
//
// It is decoupled from the network through the Fetcher interface, so it can be
// driven by a real HTTP client in production or an in-memory MapFetcher in tests
// and demos — deterministic and offline.
package crawler

import (
	"context"
	"sort"
	"sync"
)

// Fetcher retrieves the links found on a page. Implementations must honour ctx.
type Fetcher interface {
	Fetch(ctx context.Context, url string) (links []string, err error)
}

// Result is the outcome of visiting a single URL.
type Result struct {
	URL   string
	Depth int
	Links []string
	Err   error
}

// MapFetcher is an in-memory Fetcher backed by a static link graph, useful for
// tests, examples, and offline demos. A missing URL yields no links and no error
// (a dead end), matching how a real crawler treats an empty page.
type MapFetcher map[string][]string

// Fetch returns the links configured for url. It always respects ctx.
func (m MapFetcher) Fetch(ctx context.Context, url string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return m[url], nil
}

// Crawl visits seed and everything reachable from it up to maxDepth, following
// links breadth-first across the graph. At most workers fetches run at once; the
// same URL is never fetched twice, so cycles terminate. Cancelling ctx stops the
// crawl and returns whatever was collected so far.
//
// Results are returned sorted by (depth, URL) for a deterministic order despite
// the concurrent traversal.
func Crawl(ctx context.Context, seed string, maxDepth, workers int, f Fetcher) []Result {
	if workers < 1 {
		workers = 1
	}
	sem := make(chan struct{}, workers) // bounds concurrent fetches
	results := make(chan Result)

	var (
		mu      sync.Mutex
		visited = map[string]bool{seed: true}
		wg      sync.WaitGroup
	)

	var crawl func(url string, depth int)
	crawl = func(url string, depth int) {
		defer wg.Done()

		// Acquire a fetch slot, or bail out if cancelled.
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return
		}
		defer func() { <-sem }()

		links, err := f.Fetch(ctx, url)
		select {
		case results <- Result{URL: url, Depth: depth, Links: links, Err: err}:
		case <-ctx.Done():
			return
		}
		if err != nil || depth >= maxDepth {
			return
		}

		for _, link := range links {
			mu.Lock()
			if visited[link] {
				mu.Unlock()
				continue
			}
			visited[link] = true
			mu.Unlock()

			wg.Add(1)
			go crawl(link, depth+1)
		}
	}

	wg.Add(1)
	go crawl(seed, 0)

	// One closer goroutine closes results after every crawl goroutine exits.
	go func() {
		wg.Wait()
		close(results)
	}()

	var out []Result
	for r := range results {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return out[i].URL < out[j].URL
	})
	return out
}
