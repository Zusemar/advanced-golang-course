// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package probe

import (
	"fmt"
	"strconv"
	"testing"

	"course/w02/practice/feed"
	rangeclose "course/w02/practice/step0-range"
)

// BenchmarkMerge times every merge at 2 to 128 unbuffered inputs, with no
// sleeps and 19,968 items per op whatever the input count: the practice's
// BenchmarkMerge has 20,000 (two inputs of 10,000), and 19,968 is the nearest
// total that every input count up to 128 splits evenly (156 each at 128).
// ns/item is the cost of one item through the merge:
//
//	go test -run=X -bench=Merge -cpu=1,2,8 .
//
// A b.N loop, not b.Loop: at go1.27.1 a b.Loop benchmark takes its first
// measurement at the GOMAXPROCS left over from the last -cpu entry.
func BenchmarkMerge(b *testing.B) {
	const total = 19_968
	for _, n := range Sizes {
		per := total / n
		want := map[string]int{}
		names := make([]string, n)
		for k := range n {
			names[k] = "in" + strconv.Itoa(k)
			want[names[k]] = per
		}
		for _, m := range Merges(n) {
			b.Run(fmt.Sprintf("N=%d/%s", n, m.Name), func(b *testing.B) {
				for range b.N {
					ins := make([]<-chan feed.Item, n)
					for k := range n {
						ins[k] = feed.Source(names[k], per, 0, 0)
					}
					if err := feed.Check(rangeclose.Collect(m.Run(ins...)), want); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*total), "ns/item")
			})
		}
	}
}
