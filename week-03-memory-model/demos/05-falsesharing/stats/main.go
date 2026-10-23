// Demo 5, the cold open: two goroutines, two counters, no shared data.

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const n = 20_000_000

// Stats: hits and misses lie next to each other.
type Stats struct {
	hits   atomic.Int64
	misses atomic.Int64
}

// PaddedStats: the same two counters, 64 bytes apart.
type PaddedStats struct {
	hits   atomic.Int64
	_      [56]byte
	misses atomic.Int64
}

func run(hits, misses *atomic.Int64) time.Duration {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Go(func() {
		for range n {
			hits.Add(1)
		}
	})
	wg.Go(func() {
		for range n {
			misses.Add(1)
		}
	})
	wg.Wait()
	return time.Since(start)
}

func main() {
	var s Stats
	var p PaddedStats
	fmt.Println("Stats:      ", run(&s.hits, &s.misses).Round(time.Millisecond))
	fmt.Println("PaddedStats:", run(&p.hits, &p.misses).Round(time.Millisecond))
}
