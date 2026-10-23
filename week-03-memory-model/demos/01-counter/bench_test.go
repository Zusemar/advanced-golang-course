package main

import "testing"

func BenchmarkMutex(b *testing.B) {
	for range b.N {
		withMutex()
	}
}
func BenchmarkAtomic(b *testing.B) {
	for range b.N {
		withAtomic()
	}
}
func BenchmarkChannel(b *testing.B) {
	for range b.N {
		withOwner()
	}
}
