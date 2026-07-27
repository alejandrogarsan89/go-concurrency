// Package once provides Lazy, a generic, thread-safe memoized value computed
// exactly once on first use.
//
// It is the idiomatic answer to lazy initialization: an expensive value (a
// database connection, a parsed config, a compiled regexp) that must be built
// at most once and then shared by many goroutines. sync.Once guarantees the
// initializer runs a single time and that its writes are visible to every
// caller of Get, with no torn or partial initialization ever observable.
package once

import "sync"

// Lazy holds a value of type T that is produced on first access by an init
// function and cached for all subsequent accesses. The zero value is not
// usable; construct one with New. A Lazy must not be copied after first use.
type Lazy[T any] struct {
	once sync.Once
	init func() T
	val  T
}

// New returns a Lazy whose value will be produced by init the first time Get is
// called. init is never called until then, and is called at most once no matter
// how many goroutines call Get concurrently.
func New[T any](init func() T) *Lazy[T] {
	return &Lazy[T]{init: init}
}

// Get returns the memoized value, computing it on the first call. It is safe for
// concurrent use: concurrent first callers block until the single init call
// completes, then all observe the same value.
func (l *Lazy[T]) Get() T {
	l.once.Do(func() {
		l.val = l.init()
		l.init = nil // release the closure once we no longer need it
	})
	return l.val
}
