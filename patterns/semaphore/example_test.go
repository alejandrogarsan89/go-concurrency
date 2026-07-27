package semaphore_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/alejandrogarsan89/go-concurrency/patterns/semaphore"
)

func ExampleSemaphore() {
	// Allow at most 2 downloads to run at once, regardless of how many
	// goroutines we launch.
	sem := semaphore.New(2)
	ctx := context.Background()

	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := sem.Acquire(ctx); err != nil {
				return
			}
			defer sem.Release()
			mu.Lock()
			done++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(done)
	// Output:
	// 5
}
