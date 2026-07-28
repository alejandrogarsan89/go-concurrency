package mapreduce_test

import (
	"context"
	"math"
	"runtime"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/patterns/mapreduce"
)

// expensive simulates a CPU-bound mapper so the benchmark reflects real work.
func expensive(x int) float64 {
	acc := 0.0
	for i := 0; i < 200; i++ {
		acc += math.Sqrt(float64(x*i + 1))
	}
	return acc
}

func benchInputs(n int) []int {
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}
	return in
}

// BenchmarkSerial folds the mapped values on a single goroutine.
func BenchmarkSerial(b *testing.B) {
	inputs := benchInputs(100_000)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		acc := 0.0
		for _, x := range inputs {
			acc += expensive(x)
		}
		_ = acc
	}
}

// BenchmarkParallel does the same fold with MapReduce across all cores.
func BenchmarkParallel(b *testing.B) {
	inputs := benchInputs(100_000)
	workers := runtime.GOMAXPROCS(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = mapreduce.MapReduce(context.Background(), inputs, workers,
			expensive,
			func(a, c float64) float64 { return a + c },
			0.0,
		)
	}
}
