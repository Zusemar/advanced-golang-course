// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package sequential is step 0: Tier 0, the baseline every other step is
// measured against and checked against.
package sequential

// Run runs every task, one after the other, on the calling goroutine.
// n is ignored: this is the version with no concurrency at all.
func Run(n int, tasks []func()) {
	for _, t := range tasks {
		t()
	}
}
