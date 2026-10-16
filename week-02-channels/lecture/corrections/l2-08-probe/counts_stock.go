//go:build !l208count

package main

const counting = false

func raceCounts() (won, lost uint64) { return 0, 0 }
