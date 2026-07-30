package pubsub_test

import (
	"sync"
	"testing"

	"github.com/alejandrogarsan89/go-concurrency/apps/pubsub"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPublishFansOutToAllSubscribers(t *testing.T) {
	b := pubsub.New[int]()
	defer b.Close()

	s1, _ := b.Subscribe("news", 3)
	s2, _ := b.Subscribe("news", 3)

	for i := 1; i <= 3; i++ {
		if got := b.Publish("news", i); got != 2 {
			t.Fatalf("Publish delivered to %d subscribers, want 2", got)
		}
	}

	for _, ch := range []<-chan int{s1, s2} {
		for i := 1; i <= 3; i++ {
			if v := <-ch; v != i {
				t.Fatalf("received %d, want %d", v, i)
			}
		}
	}
}

func TestTopicsAreIsolated(t *testing.T) {
	b := pubsub.New[string]()
	defer b.Close()

	sports, _ := b.Subscribe("sports", 1)
	tech, _ := b.Subscribe("tech", 1)

	b.Publish("sports", "goal")
	if got := b.Publish("tech", "release"); got != 1 {
		t.Fatalf("tech delivered to %d, want 1", got)
	}

	if v := <-sports; v != "goal" {
		t.Fatalf("sports got %q, want goal", v)
	}
	if v := <-tech; v != "release" {
		t.Fatalf("tech got %q, want release", v)
	}
}

func TestUnsubscribeStopsDeliveryAndClosesChannel(t *testing.T) {
	b := pubsub.New[int]()
	defer b.Close()

	ch, unsub := b.Subscribe("t", 1)
	unsub()

	if got := b.Publish("t", 1); got != 0 {
		t.Fatalf("Publish after unsubscribe delivered to %d, want 0", got)
	}
	if _, ok := <-ch; ok {
		t.Fatal("channel should be closed after unsubscribe")
	}
}

func TestUnsubscribeIsIdempotent(_ *testing.T) {
	b := pubsub.New[int]()
	defer b.Close()
	_, unsub := b.Subscribe("t", 1)
	unsub()
	unsub() // must not panic (double close)
}

func TestCloseClosesAllAndRejects(t *testing.T) {
	b := pubsub.New[int]()
	ch, _ := b.Subscribe("t", 1)
	b.Close()

	if _, ok := <-ch; ok {
		t.Fatal("subscriber channel should be closed after Close")
	}
	if got := b.Publish("t", 1); got != 0 {
		t.Fatalf("Publish after Close delivered to %d, want 0", got)
	}
	// Subscribe after Close returns an already-closed channel and a no-op unsub.
	ch2, unsub := b.Subscribe("t", 1)
	if _, ok := <-ch2; ok {
		t.Fatal("Subscribe after Close should return a closed channel")
	}
	unsub()
	b.Close() // idempotent
}

func TestSlowSubscriberDropsInsteadOfBlocking(t *testing.T) {
	b := pubsub.New[int]()
	defer b.Close()

	_, _ = b.Subscribe("t", 1) // buffer 1, never drained

	if got := b.Publish("t", 1); got != 1 {
		t.Fatalf("first Publish delivered to %d, want 1", got)
	}
	// Buffer now full; the publisher must not block and the message is dropped.
	if got := b.Publish("t", 2); got != 0 {
		t.Fatalf("second Publish delivered to %d, want 0 (buffer full -> drop)", got)
	}
}

func TestConcurrentPublishSubscribeIsRaceFree(_ *testing.T) {
	b := pubsub.New[int]()
	defer b.Close()

	var wg sync.WaitGroup
	// Publishers.
	for p := 0; p < 8; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				b.Publish("t", i)
			}
		}()
	}
	// Subscribers churning in and out; each drains until its own unsubscribe
	// closes the channel, and we wait for that drain to finish (goleak-safe).
	for s := 0; s < 8; s++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				ch, unsub := b.Subscribe("t", 4)
				var drained sync.WaitGroup
				drained.Add(1)
				go func() {
					defer drained.Done()
					n := 0
					for range ch {
						n++
					}
					_ = n
				}()
				unsub()
				drained.Wait()
			}
		}()
	}
	wg.Wait()
}
