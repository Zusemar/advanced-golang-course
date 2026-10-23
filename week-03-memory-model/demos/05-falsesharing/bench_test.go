package falsesharing

import (
	"sync"
	"sync/atomic"
	"testing"
)

const iters = 200_000

func bench(b *testing.B, stride int, procs int) {
	// One flat backing array; counter i lives at byte offset i*stride.
	buf := make([]byte, procs*stride+128) // +128: room to align the base to a cache line
	ptr := func(i int) *int64 {
		return (*int64)(ptrAt(buf, i*stride))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		for g := 0; g < procs; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				p := ptr(g)
				for i := 0; i < iters; i++ {
					atomic.AddInt64(p, 1)
				}
			}(g)
		}
		wg.Wait()
	}
}

func BenchmarkStride8(b *testing.B)   { bench(b, 8, 2) }
func BenchmarkStride16(b *testing.B)  { bench(b, 16, 2) }
func BenchmarkStride32(b *testing.B)  { bench(b, 32, 2) }
func BenchmarkStride64(b *testing.B)  { bench(b, 64, 2) }
func BenchmarkStride128(b *testing.B) { bench(b, 128, 2) }
func BenchmarkStride256(b *testing.B) { bench(b, 256, 2) }
