// Package group implements an errgroup-style Group: run a collection of
// goroutines, wait for them all, and collect the first error — optionally
// cancelling the rest as soon as one fails.
//
// It is the pattern for "do these N things concurrently, but abort everything if
// any one fails". It is built entirely from primitives explained elsewhere in
// this repo: a sync.WaitGroup to join the goroutines, a sync.Once so the first
// error (and the cancellation it triggers) wins exactly once, and a
// context.Context so siblings observe the cancellation cooperatively.
package group

import (
	"context"
	"sync"
)

// Group runs a set of goroutines and waits for them to finish. The zero value is
// a valid Group that waits but never cancels; use WithContext to get a Group
// that cancels its siblings on the first error. A Group must not be copied after
// first use.
type Group struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
	cancel  context.CancelFunc
}

// WithContext returns a new Group and a derived Context. The Context is cancelled
// the first time a function passed to Go returns a non-nil error, or when Wait
// returns — whichever comes first. Pass the returned Context to the work so it
// can stop early.
func WithContext(parent context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(parent)
	return &Group{cancel: cancel}, ctx
}

// Go runs fn in a new goroutine. The first call that returns a non-nil error
// records that error and, if the Group was created with WithContext, cancels the
// Context to signal the other goroutines to stop.
func (g *Group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

// Wait blocks until all goroutines started with Go have returned, then returns
// the first non-nil error (if any). If the Group was created with WithContext,
// its Context is cancelled before Wait returns, releasing its resources.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
