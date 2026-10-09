// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package unbounded is step 1: one goroutine per task, and a WaitGroup to
// wait for all of them.
package unbounded

import "sync"

// Run starts a goroutine for every task and returns when all have finished.
// n is ignored, which is the bug this step exists to show: a million tasks
// means a million goroutines alive at once.
func Run(n int, tasks []func()) {
	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Go(t) // Go 1.25: wg.Add(1); go func() { defer wg.Done(); t() }()
	}
	wg.Wait()
}
