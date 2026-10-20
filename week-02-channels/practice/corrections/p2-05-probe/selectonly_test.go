// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package probe

import (
	"fmt"
	"testing"

	"course/w02/practice/feed"
)

// BenchmarkSelectOnly times one select over n channels that are all ready,
// with the merge's n-case shape and nothing else: one goroutine, no sender to
// wake and no receiver to hand off to. Every channel has a buffer of one and
// holds one item; the chosen item is put back. ns/op is one select plus that
// one buffered send, so its growth with n is what n cases cost the select
// itself:
//
//	go test -run=X -bench=SelectOnly -cpu=1 .
func BenchmarkSelectOnly(b *testing.B) {
	readies := map[int]func([]chan feed.Item, int){
		2: ready2, 4: ready4, 8: ready8, 16: ready16, 32: ready32, 64: ready64, 128: ready128,
	}
	for _, n := range Sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			cs := make([]chan feed.Item, n)
			for k := range cs {
				cs[k] = make(chan feed.Item, 1)
				cs[k] <- feed.Item{From: "in", Seq: k}
			}
			b.ResetTimer()
			readies[n](cs, b.N)
		})
	}
}
