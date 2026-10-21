// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package practice

import (
	"runtime"
	"testing"
	"time"

	"course/w02/practice/feed"
	rangeclose "course/w02/practice/step0-range"
	goroutines "course/w02/practice/step1-goroutines"
	selectspin "course/w02/practice/step2-select-spin"
	selectnil "course/w02/practice/step3-select-nil"
	earlyexit "course/w02/practice/step4-early-exit"
	drain "course/w02/practice/step5-drain"
)

// TestSteps checks what every step delivers. Run it with -race.
func TestSteps(t *testing.T) {
	want := map[string]int{"a": 300, "b": 40}
	src := func() (<-chan feed.Item, <-chan feed.Item) {
		return feed.Source("a", 300, 0, 0), feed.Source("b", 40, 100*time.Microsecond, 4)
	}

	if err := feed.Check(rangeclose.Collect(feed.Source("a", 300, 0, 8)), map[string]int{"a": 300}); err != nil {
		t.Errorf("step 0: %v", err)
	}
	a, b := src()
	if err := feed.Check(rangeclose.Collect(goroutines.Merge(a, b)), want); err != nil {
		t.Errorf("step 1: %v", err)
	}
	var wasted int64
	a, b = src()
	if err := feed.Check(rangeclose.Collect(selectspin.Merge(a, b, &wasted)), want); err != nil {
		t.Errorf("step 2: %v", err)
	}
	if wasted < 2 {
		t.Errorf("step 2: %d receives from a closed input, want at least 2", wasted)
	}
	wasted = 0
	a, b = src()
	if err := feed.Check(rangeclose.Collect(selectnil.Merge(a, b, &wasted)), want); err != nil {
		t.Errorf("step 3: %v", err)
	}
	if wasted != 2 {
		t.Errorf("step 3: %d receives from a closed input, want exactly 2", wasted)
	}
}

// TestDrainLeavesNothing checks that step 5 leaves no goroutine behind and
// that step 4 does. Step 4's leak is permanent, so it runs last.
func TestDrainLeavesNothing(t *testing.T) {
	base := runtime.NumGoroutine()
	got := drain.FirstK(5, feed.Source("a", 100, 0, 0), feed.Source("b", 100, 0, 0))
	if err := feed.CheckPrefix(got); err != nil || len(got) != 5 {
		t.Fatalf("step 5: got %d items, %v", len(got), err)
	}
	if n := settle(base); n != base {
		t.Errorf("step 5: %d goroutines after FirstK, want %d", n, base)
	}
	got = earlyexit.FirstK(5, feed.Source("a", 100, 0, 0), feed.Source("b", 100, 0, 0))
	if err := feed.CheckPrefix(got); err != nil || len(got) != 5 {
		t.Fatalf("step 4: got %d items, %v", len(got), err)
	}
	if n := settle(base); n == base {
		t.Errorf("step 4: no goroutine leaked; the demonstration is broken")
	}
}

// settle waits up to a second for the goroutine count to fall back to base,
// and returns the last count it saw.
func settle(base int) int {
	n := runtime.NumGoroutine()
	for deadline := time.Now().Add(time.Second); n != base && time.Now().Before(deadline); {
		time.Sleep(10 * time.Millisecond)
		n = runtime.NumGoroutine()
	}
	return n
}
