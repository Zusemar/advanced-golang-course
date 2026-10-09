// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package workerpool is step 4: the same fixed pool of n workers as step 3,
// but each worker takes its next task only when it has finished the last.
package workerpool

import "sync"

// Run starts n workers that pull tasks from one shared queue until it is
// closed. The split is decided while the tasks run, so no worker idles while
// work remains.
func Run(n int, tasks []func()) {
	queue := make(chan func())
	var wg sync.WaitGroup
	for range max(n, 1) {
		wg.Go(func() {
			for t := range queue {
				t()
			}
		})
	}
	for _, t := range tasks {
		queue <- t
	}
	close(queue)
	wg.Wait()
}
