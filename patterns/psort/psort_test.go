package psort_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/psort"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func intLess(a, b int) bool { return a < b }

func isSorted(s []int) bool {
	for i := 1; i < len(s); i++ {
		if s[i] < s[i-1] {
			return false
		}
	}
	return true
}

func TestSortMatchesStdlib(t *testing.T) {
	for _, n := range []int{0, 1, 2, 100, seqSize(), 50_000} {
		got := randomInts(n)
		want := append([]int(nil), got...)
		sort.Ints(want)

		psort.Sort(got, intLess)

		if len(got) != len(want) {
			t.Fatalf("n=%d: length changed", n)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("n=%d: got[%d]=%d, want %d", n, i, got[i], want[i])
			}
		}
	}
}

func TestSortAlreadySortedAndReversed(t *testing.T) {
	n := 20_000
	asc := make([]int, n)
	desc := make([]int, n)
	for i := 0; i < n; i++ {
		asc[i] = i
		desc[i] = n - i
	}
	psort.Sort(asc, intLess)
	psort.Sort(desc, intLess)
	if !isSorted(asc) || !isSorted(desc) {
		t.Fatal("failed to sort already-sorted or reversed input")
	}
}

func TestSortIsStable(t *testing.T) {
	type kv struct{ key, ord int }
	n := 10_000
	items := make([]kv, n)
	for i := 0; i < n; i++ {
		items[i] = kv{key: i % 10, ord: i} // many equal keys
	}
	psort.Sort(items, func(a, b kv) bool { return a.key < b.key })

	// Within each key group, the original order (ord) must be preserved.
	for i := 1; i < len(items); i++ {
		if items[i].key == items[i-1].key && items[i].ord < items[i-1].ord {
			t.Fatalf("not stable at %d: keys equal but order reversed", i)
		}
	}
}

func randomInts(n int) []int {
	r := rand.New(rand.NewSource(1))
	s := make([]int, n)
	for i := range s {
		s[i] = r.Intn(1_000_000)
	}
	return s
}

// seqSize returns a size just above the sequential cutoff to exercise the
// parallel path near its boundary.
func seqSize() int { return (1 << 11) + 1 }
