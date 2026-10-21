// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package selectspin is step 2: fan-in of two inputs with one select, written
// wrong on purpose. It delivers every item, and burns a CPU doing it.
package selectspin

import "course/w02/practice/feed"

// Merge merges a and b with a single goroutine running a select loop. When an
// input closes, the loop only notes it. But a closed channel is always ready
// to receive, so select keeps choosing that case, and the loop
// spins until the other input closes too. *wasted counts those receives from
// a closed channel. Read it only after out is closed.
func Merge(a, b <-chan feed.Item, wasted *int64) <-chan feed.Item {
	out := make(chan feed.Item)
	go func() {
		defer close(out)
		aOpen, bOpen := true, true
		for aOpen || bOpen {
			select {
			case it, ok := <-a:
				if !ok {
					aOpen = false
					*wasted++
					continue
				}
				out <- it
			case it, ok := <-b:
				if !ok {
					bOpen = false
					*wasted++
					continue
				}
				out <- it
			}
		}
	}()
	return out
}
