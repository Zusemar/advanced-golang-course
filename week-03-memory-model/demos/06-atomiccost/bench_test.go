package atomiccost

import (
	"sync"
	"sync/atomic"
	"testing"
)

var plain int64
var at atomic.Int64
var mu sync.Mutex
var sink int64

func BenchmarkPlainLoad(b *testing.B) {
	for range b.N {
		sink = plain
	}
}
func BenchmarkPlainStore(b *testing.B) {
	for range b.N {
		plain = 1
	}
}
func BenchmarkAtomicLoad(b *testing.B) {
	for range b.N {
		sink = at.Load()
	}
}
func BenchmarkAtomicStore(b *testing.B) {
	for range b.N {
		at.Store(1)
	}
}
func BenchmarkAtomicAdd(b *testing.B) {
	for range b.N {
		at.Add(1)
	}
}
func BenchmarkAtomicCAS(b *testing.B) {
	for range b.N {
		at.CompareAndSwap(0, 0)
	}
}

func BenchmarkMutexLockUnlock(b *testing.B) {
	for range b.N {
		mu.Lock()
		mu.Unlock()
	}
}

// Contended: every P hammers the same word.
func BenchmarkAtomicAddShared(b *testing.B) {
	var c atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
}

func BenchmarkMutexShared(b *testing.B) {
	var m sync.Mutex
	var c int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			c++
			m.Unlock()
		}
	})
	_ = c
}

func BenchmarkAtomicLoadShared(b *testing.B) {
	var c atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		var s int64
		for pb.Next() {
			s += c.Load()
		}
		sink = s
	})
}
