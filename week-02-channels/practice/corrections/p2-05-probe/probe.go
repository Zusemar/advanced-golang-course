// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package probe asks the question from the practice: does step 1 (a goroutine
// per input) stay faster than step 3 (one select loop) as the number of inputs
// grows? Step 3 takes exactly two inputs, so this package generalises it to 2,
// 4, 8, 16, 32, 64 and 128 (sel2.go to sel128.go, written by gen.go), and adds
// the same loop over reflect.Select, which takes any number of inputs at run
// time.
//
// Every merge here is timed on the same total number of items, split evenly
// across its inputs. See cmd/merge for the timing and bench_test.go for the
// same comparison through go test.
package probe

//go:generate go run gen.go

import (
	"course/w02/practice/feed"
	goroutines "course/w02/practice/step1-goroutines"
	selectnil "course/w02/practice/step3-select-nil"
)

// Sizes are the input counts that have a generated select loop.
var Sizes = []int{2, 4, 8, 16, 32, 64, 128}

// A Merge is one way to merge inputs, under the name the driver and the
// benchmarks print.
type Merge struct {
	Name string
	Run  func(ins ...<-chan feed.Item) <-chan feed.Item
}

// Merges returns every merge that takes n inputs: step 1 itself, the generated
// select loop, the reflect.Select loop and, for n = 2 only, step 3 itself. Step
// 3 is the control for the generator: at two inputs the generated loop differs
// from it only in how it counts the inputs still open, so the two must time
// the same.
func Merges(n int) []Merge {
	ms := []Merge{{"goroutines", goroutines.Merge}}
	if sel := selects[n]; sel != nil {
		ms = append(ms, Merge{"select", sel})
	}
	ms = append(ms, Merge{"reflect", ReflectSelect})
	if n == 2 {
		ms = append(ms, Merge{"step3", func(ins ...<-chan feed.Item) <-chan feed.Item {
			var wasted int64
			return selectnil.Merge(ins[0], ins[1], &wasted)
		}})
	}
	return ms
}

var selects = map[int]func(ins ...<-chan feed.Item) <-chan feed.Item{
	2:   Select2,
	4:   Select4,
	8:   Select8,
	16:  Select16,
	32:  Select32,
	64:  Select64,
	128: Select128,
}
