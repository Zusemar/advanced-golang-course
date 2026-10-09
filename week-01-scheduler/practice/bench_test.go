// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package practice

import (
	"runtime"
	"testing"

	"course/w01/practice/work"
)

// BenchmarkRun times every step with n = GOMAXPROCS, so -cpu=1,2,8 sets both
// the number of Ps and the executor's bound:
//
//	go test -run=X -bench=Run -benchtime=5x -cpu=1,2,8 .
//
// Building the 1000 task closures is inside the timed loop. It costs tens of
// microseconds against a run of tens of milliseconds.
//
// This is a b.N loop, not b.Loop, on purpose. At go1.27.1 a b.Loop benchmark
// records its first measurement before `go test` sets GOMAXPROCS for the first
// -cpu value (testing/benchmark.go: run1 runs before processBench, and launch
// does not re-run a b.Loop benchmark). With -cpu=1,2,8 that first "-cpu=1"
// figure was taken at 8 Ps and showed a 7x speedup for one core. Do not use
// -benchtime=1x either: that reuses the same early measurement.
func BenchmarkRun(b *testing.B) {
	for _, s := range Steps {
		b.Run(s.Name, func(b *testing.B) {
			n := runtime.GOMAXPROCS(0)
			for range b.N {
				l, err := work.New(1000, "skewed")
				if err != nil {
					b.Fatal(err)
				}
				s.Run(n, l.Tasks())
				if err := l.Check(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestSteps is the correctness oracle for every step, at several bounds and
// both shapes: every task runs exactly once, and no bounded step ever has more
// than n tasks running at once. Run it with -race.
func TestSteps(t *testing.T) {
	for _, s := range Steps {
		for _, n := range []int{1, 2, 3, 8} {
			for _, shape := range []string{"skewed", "even"} {
				l, err := work.New(97, shape)
				if err != nil {
					t.Fatal(err)
				}
				s.Run(n, l.Tasks())
				if err := l.Check(); err != nil {
					t.Errorf("%s n=%d %s: %v", s.Name, n, shape, err)
				}
				if s.Name != "1-unbounded" && l.PeakRunning() > n {
					t.Errorf("%s n=%d %s: %d tasks ran at once", s.Name, n, shape, l.PeakRunning())
				}
			}
		}
	}
}
