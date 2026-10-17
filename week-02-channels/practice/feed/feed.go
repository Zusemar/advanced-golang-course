// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package feed is the practice session's plumbing: sources that send numbered
// items and close, and a check that a merge delivered every item exactly once
// and in order per source. Nothing here is part of the lesson. Read the step
// packages instead.
package feed

import (
	"fmt"
	"time"
)

// Item is one value on a channel: which source sent it, and its sequence
// number within that source.
type Item struct {
	From string
	Seq  int
}

// Source starts a goroutine that sends n items on a new channel with buffer
// capacity buf, sleeping gap before each send, and then closes the channel.
// The goroutine owns the channel: it is the only sender and the only closer.
func Source(name string, n int, gap time.Duration, buf int) <-chan Item {
	ch := make(chan Item, buf)
	go func() {
		defer close(ch)
		for i := range n {
			if gap > 0 {
				time.Sleep(gap)
			}
			ch <- Item{From: name, Seq: i}
		}
	}()
	return ch
}

// Check reports whether got holds exactly want[name] items from each source,
// each once, in order within its source. Items from different sources may
// interleave in any way: a merge promises no order between inputs.
func Check(got []Item, want map[string]int) error {
	next := map[string]int{}
	for _, it := range got {
		n, ok := want[it.From]
		if !ok {
			return fmt.Errorf("item from unknown source %q", it.From)
		}
		if it.Seq != next[it.From] || it.Seq >= n {
			return fmt.Errorf("source %s: got seq %d, want %d", it.From, it.Seq, next[it.From])
		}
		next[it.From]++
	}
	for name, n := range want {
		if next[name] != n {
			return fmt.Errorf("source %s: got %d items, want %d", name, next[name], n)
		}
	}
	return nil
}

// CheckPrefix reports whether got is a valid start of a merge: from each
// source, items 0, 1, 2 ... in order, with none skipped or repeated.
func CheckPrefix(got []Item) error {
	next := map[string]int{}
	for _, it := range got {
		if it.Seq != next[it.From] {
			return fmt.Errorf("source %s: got seq %d, want %d", it.From, it.Seq, next[it.From])
		}
		next[it.From]++
	}
	return nil
}
