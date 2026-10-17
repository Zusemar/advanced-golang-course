// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package goroutines is step 1: fan-in without select. One goroutine per
// input copies it onto a shared output.
package goroutines

import (
	"sync"

	"course/w02/practice/feed"
)

// Merge returns a channel that carries every item from every input and is
// closed once all inputs are closed and drained. Each input gets its own
// forwarding goroutine; many goroutines may send on one channel. Only one may
// close it, and only after every sender is done, so a separate goroutine
// waits for all of them and closes out.
func Merge(ins ...<-chan feed.Item) <-chan feed.Item {
	out := make(chan feed.Item)
	var wg sync.WaitGroup
	for _, in := range ins {
		wg.Go(func() {
			for it := range in {
				out <- it
			}
		})
	}
	go func() {
		wg.Wait()
		close(out) // delete this line and the consumer's range never ends
	}()
	return out
}
