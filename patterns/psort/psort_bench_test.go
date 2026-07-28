package psort_test

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/psort"
)

func benchData(n int) []int {
	r := rand.New(rand.NewSource(42))
	s := make([]int, n)
	for i := range s {
		s[i] = r.Int()
	}
	return s
}

// BenchmarkStdlibSort sorts with the single-threaded standard library sort.
func BenchmarkStdlibSort(b *testing.B) {
	data := benchData(1_000_000)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		s := append([]int(nil), data...)
		b.StartTimer()
		slices.Sort(s)
	}
}

// BenchmarkParallelSort sorts the same data with the parallel merge sort.
func BenchmarkParallelSort(b *testing.B) {
	data := benchData(1_000_000)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		s := append([]int(nil), data...)
		b.StartTimer()
		psort.Sort(s, func(a, c int) bool { return a < c })
	}
}
