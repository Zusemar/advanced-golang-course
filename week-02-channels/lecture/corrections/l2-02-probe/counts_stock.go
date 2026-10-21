//go:build !l202count

package main

const counting = false

func ringCounts() (writes, overfull, badindex, overwrite uint64) { return 0, 0, 0, 0 }
