package nilcase

import "testing"

// 16 cases, all live: the poll order and the lock order both have 16 entries.
func BenchmarkAllLive(b *testing.B) { run(b, false) }

// 16 cases, 15 of them nil: selectgo omits nil channels from both orders
// (src/runtime/select.go:174), so this is a 1-case select at runtime.
func BenchmarkFifteenNil(b *testing.B) { run(b, true) }

func run(b *testing.B, nilOut bool) {
	chs := make([]chan int, 16)
	for i := range chs {
		chs[i] = make(chan int, 1)
	}
	live := chs[15]
	if nilOut {
		for i := 0; i < 15; i++ {
			chs[i] = nil
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		live <- i
		select {
		case <-chs[0]:
		case <-chs[1]:
		case <-chs[2]:
		case <-chs[3]:
		case <-chs[4]:
		case <-chs[5]:
		case <-chs[6]:
		case <-chs[7]:
		case <-chs[8]:
		case <-chs[9]:
		case <-chs[10]:
		case <-chs[11]:
		case <-chs[12]:
		case <-chs[13]:
		case <-chs[14]:
		case <-chs[15]:
		}
	}
}
