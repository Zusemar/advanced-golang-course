//go:build !race

// SPDX-License-Identifier: Apache-2.0
//
// The public timing test, kept out of -race builds by the line above: -race
// slows code by 2x to 20x, and the grader times this check without it. Do not
// edit: every _test.go is replaced by the staff battery at grading.

package parallelmap_test

import (
	"context"
	"testing"
	"time"

	parallelmap "deepgo/week-01-parallelmap"
)

// The floor: 64 calls that each wait 10ms, 8 workers. Serial is 640ms; the
// grader draws other shapes and needs half the ideal speedup, best of 3 (see
// README.md). It measures waiting, not CPU, so it does not depend on your
// core count.
func TestFloor(t *testing.T) {
	var d time.Duration
	for range 3 {
		start := time.Now()
		parallelmap.Map(context.Background(), seq(64), 8, func(_ context.Context, i int) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return i, nil
		})
		if d = time.Since(start); d < 160*time.Millisecond {
			return
		}
	}
	t.Fatalf("took %v, want under 160ms (serial would be 640ms): something is serialising the calls", d)
}
