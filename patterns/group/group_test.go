package group_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alejandrogarsan89/go-concurrency/patterns/group"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestWaitAllSucceed(t *testing.T) {
	var g group.Group
	var count int64
	for i := 0; i < 10; i++ {
		g.Go(func() error {
			atomic.AddInt64(&count, 1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait() = %v, want nil", err)
	}
	if count != 10 {
		t.Fatalf("ran %d tasks, want 10", count)
	}
}

func TestWaitReturnsFirstError(t *testing.T) {
	errBoom := errors.New("boom")
	var g group.Group
	g.Go(func() error { return nil })
	g.Go(func() error { return errBoom })
	g.Go(func() error { return nil })

	if err := g.Wait(); !errors.Is(err, errBoom) {
		t.Fatalf("Wait() = %v, want %v", err, errBoom)
	}
}

func TestWithContextCancelsSiblingsOnError(t *testing.T) {
	errBoom := errors.New("boom")
	g, ctx := group.WithContext(context.Background())

	// A sibling that fails immediately.
	g.Go(func() error { return errBoom })

	// A sibling that should be cancelled by the failure.
	cancelled := make(chan struct{})
	g.Go(func() error {
		select {
		case <-ctx.Done():
			close(cancelled)
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return errors.New("was not cancelled")
		}
	})

	if err := g.Wait(); !errors.Is(err, errBoom) {
		t.Fatalf("Wait() = %v, want %v", err, errBoom)
	}
	select {
	case <-cancelled:
	default:
		t.Fatal("sibling was not cancelled on first error")
	}
}

func TestWithContextCancelsOnWait(t *testing.T) {
	g, ctx := group.WithContext(context.Background())
	g.Go(func() error { return nil })
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait() = %v, want nil", err)
	}
	if ctx.Err() == nil {
		t.Fatal("context should be cancelled after Wait returns")
	}
}
