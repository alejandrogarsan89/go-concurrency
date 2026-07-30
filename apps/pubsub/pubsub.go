// Package pubsub implements a generic, in-memory publish/subscribe broker: many
// subscribers register interest in a topic and each receives its own copy of
// every message published to it, until it unsubscribes or the broker closes.
//
// It is a compact showcase of correct channel discipline. Each subscriber owns a
// buffered channel; the broker is the sole sender and therefore the sole closer,
// and an RWMutex makes publishing (many concurrent senders) mutually exclusive
// with subscribe/unsubscribe/close (which close channels) — so a value is never
// sent on a closed channel and no channel is closed twice. Because delivery is
// synchronous under a read lock, the broker runs no goroutines of its own and
// cannot leak any.
package pubsub

import "sync"

// Broker routes messages of type T from publishers to per-topic subscribers.
// The zero value is not usable; construct one with New.
type Broker[T any] struct {
	mu     sync.RWMutex
	topics map[string]map[*subscriber[T]]struct{}
	closed bool
}

type subscriber[T any] struct {
	ch chan T
}

// New returns an empty, ready-to-use Broker.
func New[T any]() *Broker[T] {
	return &Broker[T]{topics: make(map[string]map[*subscriber[T]]struct{})}
}

// Subscribe registers interest in topic and returns a channel that receives
// messages published to it, plus an unsubscribe function. The channel is buffered
// to buffer messages (buffer is clamped to at least 0); a subscriber that does
// not keep up may miss messages sent while its buffer is full (see Publish).
//
// The returned unsubscribe function removes the subscription and closes its
// channel; it is idempotent. If the broker is already closed, Subscribe returns a
// closed channel and a no-op unsubscribe.
func (b *Broker[T]) Subscribe(topic string, buffer int) (<-chan T, func()) {
	if buffer < 0 {
		buffer = 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan T)
		close(ch)
		return ch, func() {}
	}

	sub := &subscriber[T]{ch: make(chan T, buffer)}
	subs := b.topics[topic]
	if subs == nil {
		subs = make(map[*subscriber[T]]struct{})
		b.topics[topic] = subs
	}
	subs[sub] = struct{}{}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			if set, ok := b.topics[topic]; ok {
				if _, ok := set[sub]; ok {
					delete(set, sub)
					close(sub.ch)
					if len(set) == 0 {
						delete(b.topics, topic)
					}
				}
			}
		})
	}
	return sub.ch, unsubscribe
}

// Publish delivers msg to every current subscriber of topic. Delivery is
// non-blocking per subscriber: if a subscriber's buffer is full, the message is
// dropped for that subscriber rather than blocking the publisher or other
// subscribers. It returns the number of subscribers that received the message.
func (b *Broker[T]) Publish(topic string, msg T) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return 0
	}
	delivered := 0
	for sub := range b.topics[topic] {
		select {
		case sub.ch <- msg:
			delivered++
		default: // subscriber not keeping up; drop for it
		}
	}
	return delivered
}

// Close shuts the broker down: it closes every subscriber channel (so ranging
// subscribers exit cleanly) and rejects further Publish and Subscribe calls. It
// is idempotent.
func (b *Broker[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for topic, subs := range b.topics {
		for sub := range subs {
			close(sub.ch)
		}
		delete(b.topics, topic)
	}
}
