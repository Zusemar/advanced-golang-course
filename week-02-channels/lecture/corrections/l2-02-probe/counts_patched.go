//go:build l202count

package main

import "runtime"

const counting = true

// ringCounts exists only in a runtime patched with guard.diff.
func ringCounts() (writes, overfull, badindex, overwrite uint64) { return runtime.RingRaceCounts() }
