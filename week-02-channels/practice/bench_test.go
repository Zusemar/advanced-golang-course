// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package practice

import (
	"testing"

	"course/w02/practice/feed"
	rangeclose "course/w02/practice/step0-range"
	goroutines "course/w02/practice/step1-goroutines"
	selectnil "course/w02/practice/step3-select-nil"
)

// BenchmarkMerge times step 1 against step 3 on two unbuffered inputs of
// 10,000 items each, with no sleeps; ns/item is the cost of one item through
// the merge:
//
//	go test -run=X -bench=Merge -cpu=1,2,8 .
//
// A b.N loop, not b.Loop: at go1.27.1 a b.Loop benchmark takes its first
// measurement at the GOMAXPROCS left over from the last -cpu entry.
func BenchmarkMerge(b *testing.B) {
	const n = 10_000
	want := map[string]int{"a": n, "b": n}
	merges := []struct {
		name string
		run  func(a, b <-chan feed.Item) <-chan feed.Item
	}{
		{"1-goroutines", func(a, b <-chan feed.Item) <-chan feed.Item { return goroutines.Merge(a, b) }},
		{"3-select-nil", func(a, b <-chan feed.Item) <-chan feed.Item {
			var w int64
			return selectnil.Merge(a, b, &w)
		}},
	}
	for _, m := range merges {
		b.Run(m.name, func(b *testing.B) {
			for range b.N {
				got := rangeclose.Collect(m.run(feed.Source("a", n, 0, 0), feed.Source("b", n, 0, 0)))
				if err := feed.Check(got, want); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*2*n), "ns/item")
		})
	}
}
