package crawler_test

import (
	"context"
	"fmt"

	"github.com/alejandrogarsan89/go-concurrency/apps/crawler"
)

func ExampleCrawl() {
	// An in-memory link graph — no network needed.
	web := crawler.MapFetcher{
		"/":         {"/about", "/products"},
		"/about":    {},
		"/products": {"/products/1"},
	}

	results := crawler.Crawl(context.Background(), "/", 5, 2, web)
	for _, r := range results {
		fmt.Printf("depth %d: %s\n", r.Depth, r.URL)
	}
	// Output:
	// depth 0: /
	// depth 1: /about
	// depth 1: /products
	// depth 2: /products/1
}
