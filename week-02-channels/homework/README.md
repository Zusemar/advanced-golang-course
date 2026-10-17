# Week 2 homework: Broadcaster

Due Saturday 24 October 2026, 23:59 MSK. Whatever is on `main` in your fork at that
moment is what gets graded.

## What you implement

One publisher, any number of subscribers. Every subscriber receives every message published
after it subscribed. Subscribing and unsubscribing are safe at any moment, concurrently with
publishing.

```go
func New[T any]() *Broadcaster[T]
func (b *Broadcaster[T]) Subscribe(buffer int) *Subscription[T]
func (b *Broadcaster[T]) Publish(v T)
func (b *Broadcaster[T]) Close()
func (s *Subscription[T]) C() <-chan T
func (s *Subscription[T]) Unsubscribe()

const SlowSubscriber = Block // or DropNewest, DropOldest, Disconnect: your choice
```

The contract is the doc comments in [`broadcaster.go`](broadcaster.go). **Where they and
this README disagree, the doc comments win.** Read them all: every boundary is written down
there, and the grader checks each one.

## The decision you must make: the slow subscriber

A subscriber that stops reading fills its buffer. Then `Publish` has to do something about
it, and you choose what, by setting `SlowSubscriber`:

| Policy | The slow subscriber | The publisher and everyone else |
|---|---|---|
| `Block` | gets every message | `Publish` waits for it, so everyone waits |
| `DropNewest` | gets the first `buffer` messages after it stopped, then some of the later ones, in order | never wait |
| `DropOldest` | loses its oldest undelivered messages, keeps the newest | never wait |
| `Disconnect` | gets everything up to the disconnect, with no gaps, then its channel is closed | never wait |

**Defend your choice in the comment above `SlowSubscriber`.** Say what it protects, what it
costs the publisher, the other subscribers and the slow subscriber, and why that is the right
trade. Write **at least 25 words of your own defending the policy**, and leave no `TODO`. A
missing or placeholder defence is flagged for staff review; staff read every defence either
way. The grader tests the policy you
declare, so declare the one you built.

## Tier 0: the first version

One mutex around the subscriber set, and `Publish` sends to every subscriber while holding
it, policy `Block`. It is correct as long as everyone keeps up. You write it during the
practice session, on your own machine; the instructor does not type it. It passes every
public test except `TestSlowSubscriber`. The grader keeps a copy,
and that copy fails the stall checks below. That is how we know those checks work.

## Tier 1: the real one (mandatory, about 2 hours)

**Acceptance**, all from the doc comments:

- A subscriber that keeps up gets every message whose `Publish` starts after `Subscribe`
  returns, in order, each exactly once. It never gets one whose `Publish` returned before
  `Subscribe` was called.
- `Subscribe(buffer)` panics for `buffer < 1`. A subscriber has fallen behind once `buffer`
  messages are waiting for it. Your policy applies at the earliest then, and at the latest
  when `2*buffer+1` are waiting, so memory per subscriber is bounded, however slow it is.
- Delivery is prompt: a message waiting for a subscriber that is reading reaches it within
  **100 ms**, and within **25 ms on average** across a run of messages. A design that polls
  every few milliseconds passes, but its tests run slowly: every message waits for the
  next poll.
- A subscriber that has never fallen behind loses nothing when it ends. Every message
  published to it before `Unsubscribe` or `Close` returns has been received, or is waiting
  in `C` when `C` is closed. So if you stage messages anywhere besides `C`, flush them.
- Your declared policy holds, as described in the table above. Under `DropNewest` and
  `DropOldest`, a subscriber that catches up gets new messages again. Under `Disconnect`, its
  channel is closed.
- When `Unsubscribe` returns, `C` is closed, and nothing published after that arrives.
  `Unsubscribe` is safe from any goroutine, including the reader, and safe to call twice.
- When `Close` returns, every `C` is closed, and no goroutine the `Broadcaster` started is
  still running. A leaked goroutine is a **hard fail**.
- The number of goroutines may grow with the number of subscribers, never with the number of
  messages.
- Standard library only. `sync.WaitGroup.Go` is fine.

**The stall checks: slow subscribers must not stop the world.** Subscribers that stop reading
may stop only one thing: under `Block`, the publisher. Everything else must go on. The grader
stalls several subscribers at once, and allows **100 ms** for each of these:

- Under `DropNewest`, `DropOldest` and `Disconnect`, each `Publish` call. It never waits for
  a slow subscriber, however many there are.
- Under every policy, `Subscribe` and `Unsubscribe`, while `Publish` is running past or
  blocked on slow subscribers. They must never wait for it: eight goroutines each doing
  twenty `Subscribe`/`Unsubscribe` pairs in a row must all finish within the 100 ms.
- Under `Block`, releasing a blocked `Publish` by unsubscribing the slow subscribers.

This is what the week-2 lecture was about. A send that blocks parks the sender until a
receiver arrives, so **a lock held across a send that can block** makes everyone who needs
that lock wait for the slowest subscriber. A lock around your bookkeeping is fine, as long as
nothing waits while holding it.

## Tier 2: optional, pass/fail, no deadline

Same contract, plus **no mutex anywhere: channels and atomics only.** Your package may import
`sync/atomic` but nothing else from `sync`: no `Mutex`, `RWMutex`, `Once`, `WaitGroup` or
`Cond`. Staff check it on request by listing your package's imports, and it must still pass
everything in Tier 1.

Tier 2 is an exercise, not advice. A mutex around bookkeeping is often the clearer design.

## Tests

`go version` in this directory should print `go1.27.1`. If it doesn't, see the
[practice setup](../practice/README.md#setup).

```sh
go vet .
go test -race -count=20 -cpu=1,2,8 .
go test .
```

`TestSlowSubscriber` is in `public_norace_test.go`, which is built only without `-race`, so
the `-race` command leaves it out and `go test .` runs it, the way the grader runs its
stall checks. **Run both:** an `ok` from the `-race` command alone says nothing about stalls, and
the Tier-0 broadcaster gets one.

The public tests are a short version of ours. Ours run under `-race` at `-cpu=1,2,8`, 20
times over, draw subscriber counts, buffer sizes (1 to 64) and message counts at random on
every run, and check for a leaked goroutine after every test. The stall checks run once,
without `-race`.

## Submitting

Push your solution to `main` in the fork you made for week 1. To get this week's material
into it, click **Sync fork** on GitHub. No fork yet? Fork this repository, push your
solution to `main` there, and send the link to the course chat once; later weeks are graded
from the same fork. A fork of a public repository is public, so others can see your code.
