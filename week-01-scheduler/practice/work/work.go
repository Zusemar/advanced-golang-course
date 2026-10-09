// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package work is the practice session's workload: CPU-bound tasks that touch
// no shared memory, plus the instruments every step reports. Nothing here is
// part of the lesson. Read the step packages instead.
package work

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

// unitIters is the number of xorshift rounds in one unit of work. One unit is
// roughly half a microsecond on a current desktop core.
const unitIters = 500

// Spin burns units of pure CPU time and returns a value that depends on every
// round, so the compiler cannot delete the loop.
func Spin(units int) uint64 {
	x := uint64(units) | 1
	for range units * unitIters {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
	}
	return x
}

// Load is one batch of tasks and what was observed while they ran.
type Load struct {
	tasks       []func()
	ran         []int32
	sink        []uint64
	running     atomic.Int64
	peakRunning atomic.Int64
	peakGs      atomic.Int64
}

// New builds count tasks. With shape "skewed" task i costs i+1 units, so the
// last task is count times the first. With shape "even" every task costs the
// same, and the total is the same as skewed's.
func New(count int, shape string) (*Load, error) {
	l := &Load{
		tasks: make([]func(), count),
		ran:   make([]int32, count),
		sink:  make([]uint64, count),
	}
	for i := range count {
		var units int
		switch shape {
		case "skewed":
			units = i + 1
		case "even":
			units = (count + 1) / 2
		default:
			return nil, fmt.Errorf("unknown shape %q: want skewed or even", shape)
		}
		l.tasks[i] = func() {
			raise(&l.peakRunning, l.running.Add(1))
			raise(&l.peakGs, int64(runtime.NumGoroutine()))
			l.sink[i] = Spin(units)
			l.ran[i]++
			l.running.Add(-1)
		}
	}
	return l, nil
}

// Tasks returns the tasks to hand to an executor.
func (l *Load) Tasks() []func() { return l.tasks }

// PeakRunning is the largest number of tasks seen running at the same moment.
func (l *Load) PeakRunning() int { return int(l.peakRunning.Load()) }

// PeakGoroutines is the largest runtime.NumGoroutine() seen at a task start.
// It counts the goroutines that exist, running or not, including main.
func (l *Load) PeakGoroutines() int { return int(l.peakGs.Load()) }

// Check reports whether every task ran exactly once. Call it only after the
// executor has returned. This is the correctness oracle: step 0 always
// passes it, and a faster step that does not is wrong, not fast.
func (l *Load) Check() error {
	for i, n := range l.ran {
		if n != 1 {
			return fmt.Errorf("task %d ran %d times, want 1", i, n)
		}
	}
	return nil
}

func raise(peak *atomic.Int64, v int64) {
	for {
		old := peak.Load()
		if v <= old || peak.CompareAndSwap(old, v) {
			return
		}
	}
}
