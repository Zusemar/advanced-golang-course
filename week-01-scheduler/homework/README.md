# Week 1 homework: ParallelMap

Due Saturday 17 October 2026, 23:59 MSK.

In the practice session we built an executor that runs tasks with at most `n` at once. Now
finish it: return results in order, stop on the first error, and respect cancellation.
Write `Map` in [`parallelmap.go`](parallelmap.go):

```go
func Map[T, R any](ctx context.Context, in []T, workers int,
	f func(context.Context, T) (R, error)) ([]R, error)
```

The doc comment above `Map` is the full contract, edge cases included. Budget about two
hours.

## What we check

- `out[i]` is the result for `in[i]`, and `f` runs once per element.
- At most `workers` calls run at once, and while work is left, `workers` calls are running.
  A mutex around `f` fails this. So does splitting `in` into fixed chunks up front: when
  one chunk is slow, the other workers sit idle.
- The number of goroutines depends on `workers`, not on `len(in)`. A goroutine per element
  fails, even if each one waits on a semaphore.
- The first error cancels the other calls, no new calls start, and `Map` returns `nil` and
  that error within 100 ms. Cancelling `ctx` stops `Map` the same way.
- When `Map` returns, nothing it started is still running.
- No data races.

Use the standard library only. Keep the package name, the signature and `go.mod` as they
are.

## Tests

`go version` in this directory should print `go1.27.1`. If it doesn't, see the
[practice setup](../practice/README.md#setup).

```sh
go vet .
go test -race -count=20 -cpu=1,2,8 .
go test .
```

Run both test commands. `TestFloor` checks speed, so it runs only without `-race`. A plain
`for` loop passes everything else, so a pass from `-race` alone proves little.

The public tests are a subset of ours. Ours run each check many times, with random sizes,
under `-race` and at several `GOMAXPROCS` values. They also check things the public tests
don't: the number of goroutines, that `f` gets a context derived from `ctx`, and that `Map`
waits for calls that ignore their context.

## Submitting

Fork this repository and push your solution to `main` in your fork. Send the link to the
course chat once; later weeks are graded from the same fork. A fork of a public repository
is public, so others can see your code. To get next week's material, click **Sync fork** on
GitHub.

## Optional: no per-element allocations

Make the number of allocations independent of `len(in)`. Check it with
`testing.AllocsPerRun` at 16 and at 16,384 elements. Store `Map`'s result in a
package-level variable, or the compiler may optimise the call away and you will measure
nothing. Ask in the chat if you want us to check it.
