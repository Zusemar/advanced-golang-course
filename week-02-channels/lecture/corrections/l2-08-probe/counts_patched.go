//go:build l208count

package main

import "runtime"

const counting = true

// raceCounts exists only in a runtime patched with count.diff or nocas.diff.
func raceCounts() (won, lost uint64) { return runtime.SelectRaceCounts() }
