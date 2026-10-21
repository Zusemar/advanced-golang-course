// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package selectnil is step 3: step 2's select loop, with each closed input
// switched off by setting its variable to nil.
package selectnil

import "course/w02/practice/feed"

// Merge merges a and b like step 2. A receive from a nil channel is never
// ready, so select leaves a nil case out entirely. Setting a
// closed input to nil switches off its case without changing the select.
// *wasted counts receives from a closed channel, as in step 2: here each
// input is received from after closing exactly once.
func Merge(a, b <-chan feed.Item, wasted *int64) <-chan feed.Item {
	out := make(chan feed.Item)
	go func() {
		defer close(out)
		for a != nil || b != nil {
			select {
			case it, ok := <-a:
				if !ok {
					a = nil
					*wasted++
					continue
				}
				out <- it
			case it, ok := <-b:
				if !ok {
					b = nil
					*wasted++
					continue
				}
				out <- it
			}
		}
	}()
	return out
}
