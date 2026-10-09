// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package chunked is step 3: a fixed pool of n workers, built from nothing
// but goroutines and a WaitGroup. Each worker owns one contiguous chunk.
package chunked

import "sync"

// Run splits tasks into n contiguous chunks and gives each chunk to its own
// goroutine. Exactly n goroutines, no channel, no lock. The split is decided
// before any task runs, so a worker that finishes early has nothing to take.
func Run(n int, tasks []func()) {
	n = max(n, 1)
	var wg sync.WaitGroup
	for w := range n {
		lo, hi := w*len(tasks)/n, (w+1)*len(tasks)/n
		wg.Go(func() {
			for _, t := range tasks[lo:hi] {
				t()
			}
		})
	}
	wg.Wait()
}
