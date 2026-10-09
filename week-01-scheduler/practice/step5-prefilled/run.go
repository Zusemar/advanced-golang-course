// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package prefilled is step 5: step 4's worker pool with the queue filled
// before any worker starts, so no goroutine has to feed it while tasks run.
package prefilled

import "sync"

// Run puts every task in a queue big enough to hold them all, closes it, and
// only then starts n workers to drain it. Compared with step 4 the loop moved
// above the workers and the channel got a buffer; nothing else changed.
func Run(n int, tasks []func()) {
	queue := make(chan func(), len(tasks))
	for _, t := range tasks {
		queue <- t
	}
	close(queue)
	var wg sync.WaitGroup
	for range max(n, 1) {
		wg.Go(func() {
			for t := range queue {
				t()
			}
		})
	}
	wg.Wait()
}
