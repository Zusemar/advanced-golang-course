// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package probe

import (
	"fmt"
	"strconv"
	"testing"

	"course/w02/practice/feed"
	rangeclose "course/w02/practice/step0-range"
	goroutines "course/w02/practice/step1-goroutines"
)

// TestMerges checks that every merge delivers every item once, in order per
// source, at every input count. Run it with -race.
func TestMerges(t *testing.T) {
	for _, n := range Sizes {
		for _, m := range Merges(n) {
			t.Run(fmt.Sprintf("N=%d/%s", n, m.Name), func(t *testing.T) {
				want := map[string]int{}
				ins := make([]<-chan feed.Item, n)
				for k := range n {
					name := "in" + strconv.Itoa(k)
					want[name] = 50 + 7*k // uneven, so inputs close at different times
					ins[k] = feed.Source(name, want[name], 0, k%3)
				}
				if err := feed.Check(rangeclose.Collect(m.Run(ins...)), want); err != nil {
					t.Error(err)
				}
			})
		}
	}
}

// TestFewInputs checks the edges: with no inputs, step 1 and reflect.Select
// close their output at once; with one, they forward it. A generated loop
// takes only its own input count and panics on any other.
func TestFewInputs(t *testing.T) {
	for _, m := range []Merge{{"goroutines", goroutines.Merge}, {"reflect", ReflectSelect}} {
		if got := rangeclose.Collect(m.Run()); len(got) != 0 {
			t.Errorf("%s with no inputs: got %d items, want 0", m.Name, len(got))
		}
		got := rangeclose.Collect(m.Run(feed.Source("in0", 20, 0, 0)))
		if err := feed.Check(got, map[string]int{"in0": 20}); err != nil {
			t.Errorf("%s with one input: %v", m.Name, err)
		}
	}
	for n, sel := range selects {
		for _, k := range []int{0, 1, n - 1, n + 1} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("Select%d with %d inputs did not panic", n, k)
					}
				}()
				sel(make([]<-chan feed.Item, k)...)
			}()
		}
	}
}
