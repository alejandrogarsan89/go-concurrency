package crawler_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/apps/crawler"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// site is a small link graph shared by several tests.
var site = crawler.MapFetcher{
	"/":       {"/a", "/b"},
	"/a":      {"/a1", "/a2"},
	"/b":      {"/b1"},
	"/a1":     {"/"}, // cycle back to root
	"/a2":     {},
	"/b1":     {"/a"}, // cross link (already visited)
	"/orphan": {},
}

func urls(results []crawler.Result) map[string]int {
	m := make(map[string]int)
	for _, r := range results {
		m[r.URL]++
	}
	return m
}

func TestCrawlVisitsEachReachableOnce(t *testing.T) {
	results := crawler.Crawl(context.Background(), "/", 10, 4, site)
	got := urls(results)

	want := []string{"/", "/a", "/b", "/a1", "/a2", "/b1"}
	if len(got) != len(want) {
		t.Fatalf("visited %d urls (%v), want %d", len(got), got, len(want))
	}
	for _, u := range want {
		if got[u] != 1 {
			t.Fatalf("url %q visited %d times, want exactly 1", u, got[u])
		}
	}
	if _, ok := got["/orphan"]; ok {
		t.Fatal("unreachable /orphan should not be visited")
	}
}

func TestDepthLimit(t *testing.T) {
	results := crawler.Crawl(context.Background(), "/", 1, 4, site)
	got := urls(results)
	// depth 0: "/"; depth 1: "/a", "/b". Nothing deeper.
	want := []string{"/", "/a", "/b"}
	if len(got) != len(want) {
		t.Fatalf("with maxDepth=1 visited %v, want %v", got, want)
	}
	for _, u := range want {
		if got[u] != 1 {
			t.Fatalf("url %q not visited exactly once", u)
		}
	}
}

func TestBoundsConcurrentFetches(t *testing.T) {
	const workers = 3
	tf := &trackingFetcher{f: site}
	crawler.Crawl(context.Background(), "/", 10, workers, tf)
	if tf.max > workers {
		t.Fatalf("observed %d concurrent fetches, want <= %d", tf.max, workers)
	}
}

func TestCancellationReturnsWithoutLeak(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel up front
	// Should return promptly with at most the seed; goleak (TestMain) proves
	// no goroutine is left behind.
	_ = crawler.Crawl(ctx, "/", 10, 4, site)
}

// trackingFetcher records the peak number of concurrent Fetch calls.
type trackingFetcher struct {
	f    crawler.Fetcher
	mu   sync.Mutex
	inFl int64
	max  int64
}

func (t *trackingFetcher) Fetch(ctx context.Context, url string) ([]string, error) {
	cur := atomic.AddInt64(&t.inFl, 1)
	t.mu.Lock()
	if cur > t.max {
		t.max = cur
	}
	t.mu.Unlock()
	time.Sleep(time.Millisecond) // hold the slot so overlaps are observable
	atomic.AddInt64(&t.inFl, -1)
	return t.f.Fetch(ctx, url)
}
