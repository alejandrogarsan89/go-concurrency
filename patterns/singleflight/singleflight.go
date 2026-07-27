// Package singleflight collapses concurrent, identical calls into a single
// execution whose result is shared by all callers.
//
// It solves the "thundering herd" / "cache stampede" problem: when a cached
// value expires and a thousand goroutines simultaneously try to recompute it,
// singleflight lets exactly one do the work while the other 999 wait for and
// share that one result. It is built from a mutex-guarded map of in-flight
// calls, each tracked by a sync.WaitGroup that waiters block on.
package singleflight

import "sync"

// call is a single in-flight or completed execution shared by duplicate callers.
type call[V any] struct {
	wg       sync.WaitGroup
	val      V
	err      error
	dups     int  // number of callers that joined this in-flight call
	shared   bool // whether the result was shared with duplicate callers
	panicVal any  // non-nil if fn panicked, re-raised to every caller
}

// Group deduplicates concurrent calls keyed by K, each producing a value of type
// V. The zero value is ready to use. A Group must not be copied after first use.
type Group[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]*call[V]
}

// Do executes fn and returns its result, making sure that only one execution is
// in flight for a given key at a time. Duplicate callers for the same key while
// a call is in flight block until it completes, then receive the same value and
// error. The boolean result reports whether the value was shared with other
// concurrent callers.
//
// If fn panics, the panic is re-raised in every caller for that key (rather than
// leaving waiters blocked forever), and the key is released so later calls are
// unaffected.
//
// The key is freed as soon as fn returns, so a later Do for the same key runs
// fn again (this is a de-duplicator, not a cache).
//
//nolint:revive // signature mirrors golang.org/x/sync/singleflight: (value, error, shared).
func (g *Group[K, V]) Do(key K, fn func() (V, error)) (V, error, bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[K]*call[V])
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		g.mu.Unlock()
		c.wg.Wait() // wait for the in-flight owner to finish
		if c.panicVal != nil {
			panic(c.panicVal)
		}
		return c.val, c.err, true
	}
	c := new(call[V])
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// This goroutine owns the call: run fn exactly once.
	g.doCall(c, key, fn)
	if c.panicVal != nil {
		panic(c.panicVal)
	}
	return c.val, c.err, c.shared
}

// doCall runs fn for the owning goroutine and, whatever happens, releases the
// waiters and frees the key. Cleanup is deferred so a panic in fn cannot strand
// blocked callers or poison the key; the panic value is captured and re-raised
// by Do in both the owner and the waiters.
func (g *Group[K, V]) doCall(c *call[V], key K, fn func() (V, error)) {
	normalReturn := false
	defer func() {
		g.mu.Lock()
		c.shared = c.dups > 0
		delete(g.m, key)
		g.mu.Unlock()
		c.wg.Done()
	}()
	defer func() {
		if !normalReturn {
			// fn panicked (or called runtime.Goexit); capture a panic so waiters
			// re-raise it instead of observing a bogus zero value.
			if r := recover(); r != nil {
				c.panicVal = r
			}
		}
	}()

	c.val, c.err = fn()
	normalReturn = true
}
