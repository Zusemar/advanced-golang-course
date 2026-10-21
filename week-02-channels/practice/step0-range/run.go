// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package rangeclose is step 0: one input, read to the end. It is the
// receive side every later step is built from.
package rangeclose

import "course/w02/practice/feed"

// Collect receives from in until in is closed and returns what it received.
// range stops only when the channel is closed and its buffer is empty:
// values sent before close are still delivered after it.
func Collect(in <-chan feed.Item) []feed.Item {
	var got []feed.Item
	for it := range in {
		got = append(got, it)
	}
	return got
}
