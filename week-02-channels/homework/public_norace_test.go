//go:build !race

// SPDX-License-Identifier: Apache-2.0
//
// The public timing test, kept out of -race builds by the line above: -race
// slows code by 2x to 20x, and the grader times this check without it. Do not
// edit: every _test.go is replaced by the staff battery at grading.

package broadcaster_test

import (
	"sync/atomic"
	"testing"
	"time"

	broadcaster "deepgo/week-02-broadcaster"
)

// One subscriber stops reading. Under Block, Publish waits for it, but
// Subscribe and Unsubscribe do not, and unsubscribing it releases Publish.
// Under the other policies, Publish never waits: the grader allows 100 ms.
func TestSlowSubscriber(t *testing.T) {
	b := broadcaster.New[int]()
	stalled := b.Subscribe(1)
	within := func(what string, fn func()) {
		t.Helper()
		done := make(chan struct{})
		go func() { fn(); close(done) }()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("%s took over a second: something is waiting on the slow subscriber", what)
		}
	}
	if broadcaster.SlowSubscriber != broadcaster.Block {
		for i := range 20 {
			start := time.Now()
			within("Publish", func() { b.Publish(i) })
			if d := time.Since(start); d > 100*time.Millisecond {
				t.Fatalf("Publish took %v with a slow subscriber; the grader allows 100ms", d)
			}
		}
		b.Close()
		return
	}
	var published atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range 5 {
			b.Publish(i)
			published.Add(1)
		}
	}()
	time.Sleep(50 * time.Millisecond)
	if published.Load() == 5 {
		t.Fatal("Block: Publish never waited for a subscriber that stopped reading")
	}
	within("Subscribe while Publish is blocked", func() { b.Subscribe(1).Unsubscribe() })
	within("Unsubscribe of the slow subscriber", stalled.Unsubscribe)
	within("Publish after the slow subscriber left", func() { <-done })
	b.Close()
}
