package lockcost

import "testing"

// 16 cases on 16 distinct channels: sellock takes 16 locks.
func BenchmarkSixteenChans(b *testing.B) { run(b, false) }

// 16 cases on ONE channel: sellock skips adjacent duplicates in lock order
// (src/runtime/select.go:36-42), so it takes exactly one lock.
// The pollorder shuffle and the heap sort still run over all 16 cases.
func BenchmarkOneChanSixteenCases(b *testing.B) { run(b, true) }

func run(b *testing.B, same bool) {
	chs := make([]chan int, 16)
	if same {
		c := make(chan int, 1)
		for i := range chs {
			chs[i] = c
		}
	} else {
		for i := range chs {
			chs[i] = make(chan int, 1)
		}
	}
	live := chs[15]
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
