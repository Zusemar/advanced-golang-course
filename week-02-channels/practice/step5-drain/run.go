// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package drain is step 5: step 4 with the merged channel drained to its
// close, so every goroutine upstream can finish.
package drain

import (
	"course/w02/practice/feed"
	goroutines "course/w02/practice/step1-goroutines"
)

// FirstK returns the first k items, like step 4, and then keeps receiving and
// discarding until out is closed. Compared with step 4 one line was added.
// The price is time: FirstK now returns only when every source has finished.
// Telling the sources to stop early needs a signal travelling the other way,
// which is week 6 (context).
func FirstK(k int, ins ...<-chan feed.Item) []feed.Item {
	out := goroutines.Merge(ins...)
	var got []feed.Item
	for it := range out {
		got = append(got, it)
		if len(got) == k {
			break
		}
	}
	for range out { // drain: receive until close, keep nothing
	}
	return got
}
