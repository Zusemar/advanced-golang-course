// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package earlyexit is step 4: a consumer that stops reading before the
// merged channel is closed. It returns the right answer and leaks.
package earlyexit

import (
	"course/w02/practice/feed"
	goroutines "course/w02/practice/step1-goroutines"
)

// FirstK merges ins with step 1 and returns the first k items. Then it stops
// receiving. Every goroutine still trying to send on out, or on an input
// that a forwarder no longer reads, is blocked forever: nothing will ever
// receive from those channels again. No error, no panic, no deadlock report,
// because main is still running.
func FirstK(k int, ins ...<-chan feed.Item) []feed.Item {
	out := goroutines.Merge(ins...)
	var got []feed.Item
	for it := range out {
		got = append(got, it)
		if len(got) == k {
			break
		}
	}
	return got
}
