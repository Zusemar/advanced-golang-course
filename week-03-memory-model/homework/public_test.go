// SPDX-License-Identifier: Apache-2.0
//
// The public battery. Do not edit: it is replaced by the staff battery at
// grading, which checks the same contract more thoroughly.
//
//	go test -race -count=20 -cpu=1,2,8 .
//
// The benchmarks are for Tier 1's speed requirement. Run them WITHOUT -race,
// one -cpu value per invocation (README.md, "Measuring Tier 1"):
//
//	go test -run='^$' -bench=AddParallel -count=10 -cpu=8 .
//
// The Tier 2 Stack tests are skipped unless the test binary gets -tier2:
//
//	go test -race -count=20 -cpu=1,2,8 . -args -tier2

package shardedcounter_test

import (
	"flag"
	"sync"
	"sync/atomic"
	"testing"

	shardedcounter "deepgo/week-03-shardedcounter"
)

func TestNewIsZero(t *testing.T) {
	if got := shardedcounter.New().Sum(); got != 0 {
		t.Fatalf("New().Sum() = %d, want 0", got)
	}
}

func TestSequential(t *testing.T) {
	c := shardedcounter.New()
	var want int64
	for i := int64(-500); i <= 1000; i++ {
		c.Add(i)
		want += i
		if got := c.Sum(); got != want {
			t.Fatalf("after Add(%d): Sum() = %d, want %d", i, got, want)
		}
	}
}

func TestCountersAreIndependent(t *testing.T) {
	a, b := shardedcounter.New(), shardedcounter.New()
	a.Add(7)
	b.Add(-3)
	if got := a.Sum(); got != 7 {
		t.Fatalf("a.Sum() = %d, want 7", got)
	}
	if got := b.Sum(); got != -3 {
		t.Fatalf("b.Sum() = %d, want -3", got)
	}
}

// TestConcurrentAdd: once every Add has returned, Sum is exact.
func TestConcurrentAdd(t *testing.T) {
	const goroutines, adds = 16, 2000
	c := shardedcounter.New()
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			for i := range adds {
				if i%2 == 0 {
					c.Add(int64(g + 1))
				} else {
					c.Add(-1)
				}
			}
		})
	}
	wg.Wait()
	var want int64
	for g := range goroutines {
		want += adds / 2 * int64(g+1)
		want -= adds / 2
	}
	if got := c.Sum(); got != want {
		t.Fatalf("Sum() = %d, want %d", got, want)
	}
}

// TestSumDuringAdds: with only positive deltas, a Sum running concurrently with
// Adds counts every Add that returned before Sum was called and no Add that
// had not started when Sum returned, and successive Sums never go down.
func TestSumDuringAdds(t *testing.T) {
	const writers, adds = 8, 3000
	c := shardedcounter.New()
	var started, finished atomic.Int64
	var wg sync.WaitGroup
	for range writers {
		wg.Go(func() {
			for range adds {
				started.Add(1)
				c.Add(1)
				finished.Add(1)
			}
		})
	}
	var last int64
	for finished.Load() < writers*adds {
		lo := finished.Load()
		got := c.Sum()
		hi := started.Load()
		if got < lo || got > hi {
			t.Fatalf("Sum() = %d during Adds, want between %d and %d", got, lo, hi)
		}
		if got < last {
			t.Fatalf("Sum() = %d after an earlier Sum() = %d; positive Adds only", got, last)
		}
		last = got
	}
	wg.Wait()
	if got := c.Sum(); got != writers*adds {
		t.Fatalf("final Sum() = %d, want %d", got, writers*adds)
	}
}

// BenchmarkAddParallel is Tier 1's speed check: every goroutine adds to one
// counter. ns/op is wall time divided by the total number of Adds.
func BenchmarkAddParallel(b *testing.B) {
	c := shardedcounter.New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
	if c.Sum() <= 0 {
		b.Fatal("Sum() <= 0 after the benchmark")
	}
}

// tier0 is the Tier-0 counter, one mutex around one int64, kept here so that
// BenchmarkAddParallelTier0 measures it in the same run as yours.
type tier0 struct {
	mu sync.Mutex
	n  int64
}

func (c *tier0) Add(delta int64) {
	c.mu.Lock()
	c.n += delta
	c.mu.Unlock()
}

// BenchmarkAddParallelTier0 is the baseline Tier 1 must beat: the same loop
// against the Tier-0 counter.
func BenchmarkAddParallelTier0(b *testing.B) {
	c := &tier0{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
	if c.n <= 0 {
		b.Fatal("n <= 0 after the benchmark")
	}
}

// Tier 2 only: optional, pass/fail. README.md, "Tier 2".

var tier2 = flag.Bool("tier2", false, "also run the Tier 2 Stack tests")

func skipUnlessTier2(t *testing.T) {
	t.Helper()
	if !*tier2 {
		t.Skip("Tier 2: run with -args -tier2")
	}
}

func TestStackLIFO(t *testing.T) {
	skipUnlessTier2(t)
	s := shardedcounter.NewStack[int]()
	if v, ok := s.Pop(); ok {
		t.Fatalf("Pop() on a new stack = %d, true; want ok false", v)
	}
	for i := range 100 {
		s.Push(i)
	}
	for i := 99; i >= 0; i-- {
		if v, ok := s.Pop(); !ok || v != i {
			t.Fatalf("Pop() = %d, %v; want %d, true", v, ok, i)
		}
	}
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop() on an emptied stack: ok true, want false")
	}
}

// TestStackConcurrent: every value pushed is popped exactly once.
func TestStackConcurrent(t *testing.T) {
	skipUnlessTier2(t)
	const goroutines, per = 8, 2000
	s := shardedcounter.NewStack[int]()
	seen := make([][]int, goroutines)
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			for i := range per {
				s.Push(g*per + i)
				if v, ok := s.Pop(); ok {
					seen[g] = append(seen[g], v)
				}
			}
		})
	}
	wg.Wait()
	count := make([]int, goroutines*per)
	tally := func(v int) {
		if v < 0 || v >= len(count) {
			t.Fatalf("Pop() = %d, which was never pushed", v)
		}
		count[v]++
	}
	for _, vs := range seen {
		for _, v := range vs {
			tally(v)
		}
	}
	for v, ok := s.Pop(); ok; v, ok = s.Pop() {
		tally(v)
	}
	for v, n := range count {
		if n != 1 {
			t.Fatalf("value %d popped %d times, want 1", v, n)
		}
	}
}
