<!-- Prose: CC-BY-SA 4.0. Code in this directory: Apache-2.0. -->

# Week 3 — `ShardedCounter`

**Due Friday 30 October 2026, 23:59 MSK.** Whatever is on `main` in your repository at
that moment is what gets graded.

## What you implement

A counter that many goroutines add to at once, and that stays fast when they do.

```go
func New() *ShardedCounter
func (c *ShardedCounter) Add(delta int64)
func (c *ShardedCounter) Sum() int64
```

The contract is the doc comments in `shardedcounter.go`. **Where they and this README
disagree, the doc comments win.** `Add` and `Sum` are safe from any number of goroutines at
once. Deltas may be negative, and overflow wraps as `int64` addition does. When no `Add` runs
concurrently with it, `Sum` is exactly the total of every delta added so far. Make counters
with `New`; the zero value is not required to work.

## Tier 0: the first version

One `sync.Mutex` around one `int64`: `Add` and `Sum` each lock it. You write it during the
practice session, on your own machine; the instructor does not type it. It is correct, and
it passes every public test. The grader keeps a copy, and that copy is what Tier 1 is
measured against.

## Tier 1: the real one (mandatory, about 2 hours)

**Acceptance:**

- **Correct.** Every public test passes under `go test -race -count=20 -cpu=1,2,8 .`, run
  from inside the task directory. The grader's battery checks the same contract with more
  shapes: up to 64 writers, up to four concurrent readers, deltas of both signs. A data
  race is a fail, and so is a goroutine still running after a test: a **hard fail**.
- **Faster than Tier 0 under contention.** At `-cpu=8`, `BenchmarkAddParallel` (your
  counter) must be at least **4× faster** than `BenchmarkAddParallelTier0` (the Tier-0
  counter) in the same run: the median of Tier 0's ns/op divided by the median of yours,
  over `-count=10`. The threshold was measured and set on the course's WSL2 box, Intel
  Core i5-13600KF, 20 logical CPUs, `go1.27.1`, on 2026-10-23. Staff measure it on
  that machine, without `-race`; see "Measuring Tier 1" below. One `atomic.Int64` for the
  whole counter is correct, but it does not clear that ratio there.
- **The snapshot statement.** The doc comment on `Sum` states what `Sum` promises while
  `Add`s run concurrently with it. Replace the `TODO` in the stub; see below.
- Standard library only. A sharded counter needs little beyond `sync/atomic`, `runtime`
  and `math/rand/v2`; staff read every import, but no standard-library import fails a
  submission by itself.

### Four things from the lecture that decide the design

Lecture 3, §11:

1. **A single hot `int64` does not scale.** Every `Add` writes the same cache line, and that
   line has to travel to whichever core writes it next. More shards means more cache lines,
   which is the entire mechanism.
2. **Padding is not free.** The cache-line pad size is 64 bytes on x86 and **128 on arm64**.
   Pad each shard to its own line, and size the shard count from `GOMAXPROCS`, not a
   constant: a constant big enough for a many-core machine wastes a padded line per idle
   shard on a small one (a memory bug), and one that suits a small machine leaves a
   many-core one's goroutines contending for too few lines (a performance bug).
3. **`Sum()` has no atomic snapshot.** Reading N shards is N atomic loads. The loads are
   ordered, but `Add`s land between them, so the total may correspond to no instant in time.
4. **Atomic loads on amd64 are free until someone writes the line.** Under write load, each
   shard `Sum` loads is a line another core has just written. Under contention, what `Add`
   costs on top of its uncontended price is cache lines, and so is the cost of `Sum`.

Also decide how an `Add` picks its shard. Whatever you choose is on every `Add`'s path: if
choosing the shard writes memory that every goroutine shares, you have rebuilt the hot line
of point 1.

### The snapshot statement

`Sum` is not a snapshot, so "returns the sum" is not a contract. Write down what it *does*
promise, in the doc comment above `Sum`. The deliverable is the statement, and it must be
defensible, not just true. It should answer:

- Which `Add`s is `Sum` certain to count, and which is it certain not to count?
- What about an `Add` that runs while `Sum` is running?
- Does the result correspond to the counter's value at some instant? If not, what bound,
  if any, does it satisfy? Does your answer change when some deltas are negative?
- If one goroutine calls `Sum` twice in a row, what can it say about the two results?

**Write at least 25 words of your own** and leave no `TODO`. A missing or placeholder
statement is flagged for staff review; staff read every statement either way, against your
code.

## Measuring Tier 1

The benchmarks are in `public_test.go`. Run them **without `-race`**, and **one `-cpu` value
per `go test` invocation**, not a list such as `-cpu=1,2,8`. That is the course's
measurement rule. Lecture 3, §11 names the trap behind it: under a list, a `b.Loop`
benchmark takes its first `-count` at the wrong `GOMAXPROCS`. These two use
`b.RunParallel`, which is driven by `b.N`, so the trap does not reach them, but staff run
one `-cpu` value per invocation and so should you. From inside the task directory:

```
go test -run='^$' -bench=AddParallel -count=10 -cpu=1 .
go test -run='^$' -bench=AddParallel -count=10 -cpu=2 .
go test -run='^$' -bench=AddParallel -count=10 -cpu=8 .
```

Each line prints `BenchmarkAddParallel-N` (yours) and `BenchmarkAddParallelTier0-N` ten
times each, where `N` is the `-cpu` value; at `-cpu=1` the names carry no `-N` suffix.
Your ratio is the median of the Tier0 lines' ns/op divided by the median of yours, from the
**same** invocation; never compare figures across runs or machines. Nothing else heavy
should run while you measure: a build in another terminal is enough to move the result.

Your own laptop's ratio is a guide, not the verdict. It depends on the machine: on a laptop
with fewer than 8 cores, `-cpu=8` runs more goroutines than there are cores. The graded
figure is the staff measurement on the machine named above. `-cpu=1` is there so you can see
what sharding costs when there is no contention; it is not graded.

## Tier 2: optional, pass/fail, no deadline

A **Treiber stack** in `stack.go`: lock-free `Push` and `Pop` with `atomic.Pointer`. The stub
gives the API. Your `stack.go` must import nothing from `sync` except `sync/atomic`. It must
pass the Tier-2 tests in `public_test.go`, which skip unless the test binary is given
`-tier2`. From inside the task directory:

```
go test -race -count=20 -cpu=1,2,8 . -args -tier2
```

**And a written ABA analysis**, in `ABA.md` beside your code (lecture 3, §10). Explain why
your stack is safe from ABA in Go, what would break, and how, if you recycled nodes yourself,
and why Go has no tagged-pointer fix in `sync/atomic` at `go1.27.1`. **Name one
standard-library type that would break your argument**, and say how. The analysis is graded on
whether it explains *why the garbage collector saves you* and on that one type, not on
whether your code runs.

Tier 2 is an exercise, not advice.

## Running the tests

From inside the task directory, the one holding this README and `go.mod`, with the pinned
toolchain (`go version` in that directory should print `go1.27.1`):

```
go vet .
go test -race -count=20 -cpu=1,2,8 .
```

**How it is graded.** `public_test.go` is a short version of the grader, for you to run. After
the deadline, staff delete every `_test.go` in your directory and run a longer, private
battery in its place, under `-race` at `-cpu=1,2,8`, 20 times over, with a goroutine-leak
check after every test and shapes drawn at random on every run. Then staff run the benchmark
for the speed requirement, and read your `Sum` statement.
