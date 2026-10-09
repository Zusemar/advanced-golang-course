// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package semaphore is step 2: still one goroutine per task, but no more
// than n of them may be running a task at once.
package semaphore

import "sync"

// Run runs every task with at most n in flight. A buffered channel of
// capacity n is the counting semaphore: a send takes a slot, a receive gives
// it back. Channels are week 2; here the channel is only a counter.
func Run(n int, tasks []func()) {
	sem := make(chan struct{}, max(n, 1))
	var wg sync.WaitGroup
	for _, t := range tasks {
		sem <- struct{}{} // blocks the spawning loop while n tasks are running
		wg.Go(func() {
			defer func() { <-sem }()
			t()
		})
	}
	wg.Wait()
}
