# go-concurrency — Theory & Design Notes

This directory explains the **concurrency and parallelism** concepts behind the
library: what each pattern is, the bugs it prevents, its trade-offs, and when to
reach for it. Code lives in each `patterns/` package with full godoc; these
pages give the conceptual picture and diagrams.

## Index

The docs are organised in three layers — the theory underneath, the patterns
built on it, and the judgement to apply them well.

### Foundations — how Go concurrency actually works

| Page | Covers |
|------|--------|
| [The Go Memory Model & `happens-before`](memory-model.md) | data races, visibility, the guarantees channels/`sync`/atomics give you |
| [The Go Scheduler (G-M-P)](scheduler.md) | goroutines vs threads, `GOMAXPROCS`, work-stealing, preemption, blocking |

### Patterns — reusable building blocks

| Page | Covers |
|------|--------|
| [Fundamentals](fundamentals.md) | WaitGroup, channel generators, fan-in |
| [Worker Pools & Pipelines](pool-pipeline.md) | bounded worker pool, multi-stage pipelines |
| [Synchronization Primitives](synchronization.md) | lazy init, semaphore, errgroup, singleflight |
| [Real Parallelism & Speedup](parallelism.md) | parallel map-reduce, parallel merge sort, Amdahl, benchmarks |

### Practice — applying it correctly

| Page | Covers |
|------|--------|
| [Pitfalls & Anti-Patterns](pitfalls.md) | the classic bugs (races, leaks, deadlocks, bad closes) and their fixes |
| [Choosing the Right Primitive](decision-guide.md) | decision guide + cheat-sheet: which tool for which problem |
| [Mini-Applications](mini-apps.md) | crawler, rate limiter, pub/sub — patterns composed into real programs |

*(the pattern and practice pages together cover all five phases: fundamentals,
worker pools & pipelines, synchronization primitives, real parallelism, and mini
applications)*

## Concurrency vs Parallelism

These words are often confused; the distinction matters.

- **Concurrency** is a *structuring* concept: dealing with many things at once
  by composing independently executing tasks (goroutines). A single CPU core can
  run concurrent code by interleaving tasks.
- **Parallelism** is an *execution* concept: doing many things at the same
  instant, which requires multiple CPU cores.

> Rob Pike: *"Concurrency is about **dealing with** lots of things at once.
> Parallelism is about **doing** lots of things at once."*

Go gives you concurrency primitives (goroutines, channels, `select`); the
runtime then schedules those goroutines across `GOMAXPROCS` OS threads, turning
concurrency into parallelism when cores are available. Well-structured
concurrent code becomes parallel *for free* — but only if it's correct.

## The two hard problems this library guards against

1. **Data races** — two goroutines touch the same memory and at least one
   writes, without synchronization. The result is undefined behaviour. Go ships
   a **race detector** (`go test -race`); every test here runs under it.
2. **Goroutine leaks** — a goroutine blocks forever (usually on a channel
   send/receive) because nobody will ever unblock it. Leaks silently consume
   memory and hold resources. Every blocking goroutine here is cancellable via
   `context.Context`, and tests use [`goleak`](https://github.com/uber-go/goleak)
   to assert no goroutine outlives the test.

## Golden rules used throughout

- **The sender closes the channel, never the receiver** — and closes it exactly
  once. Closing signals "no more values"; sending on a closed channel panics.
- **Whoever starts a goroutine is responsible for it stopping.** Provide a way
  out (a closed input channel or a cancelled context).
- **`select` with `<-ctx.Done()`** on every potentially-blocking send/receive so
  a task can always be cancelled.
- **Prefer channels to share data; use mutexes to protect state.** ("Don't
  communicate by sharing memory; share memory by communicating.")
