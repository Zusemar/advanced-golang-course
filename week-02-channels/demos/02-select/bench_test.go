package sel

import "testing"

// One ready case among n. Only the number of cases varies.
func BenchmarkSelect1(b *testing.B)  { nready(b, 1) }
func BenchmarkSelect2(b *testing.B)  { nready(b, 2) }
func BenchmarkSelect4(b *testing.B)  { nready(b, 4) }
func BenchmarkSelect8(b *testing.B)  { nready(b, 8) }
func BenchmarkSelect16(b *testing.B) { nready(b, 16) }

func nready(b *testing.B, n int) {
	chs := make([]chan int, n)
	for i := range chs {
		chs[i] = make(chan int, 1)
	}
	// Only the last channel ever has a value; it is refilled each iteration.
	last := chs[n-1]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		last <- i
		selectN(chs)
	}
}

func selectN(chs []chan int) {
	switch len(chs) {
	case 1:
		select {
		case <-chs[0]:
		}
	case 2:
		select {
		case <-chs[0]:
		case <-chs[1]:
		}
	case 4:
		select {
		case <-chs[0]:
		case <-chs[1]:
		case <-chs[2]:
		case <-chs[3]:
		}
	case 8:
		select {
		case <-chs[0]:
		case <-chs[1]:
		case <-chs[2]:
		case <-chs[3]:
		case <-chs[4]:
		case <-chs[5]:
		case <-chs[6]:
		case <-chs[7]:
		}
	case 16:
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

// A plain receive, for comparison with the 1-case select above.
func BenchmarkPlainRecv(b *testing.B) {
	ch := make(chan int, 1)
	for i := 0; i < b.N; i++ {
		ch <- i
		<-ch
	}
}

// select with a default on an empty channel: compiled to selectnbrecv, no selectgo.
func BenchmarkSelectDefault(b *testing.B) {
	ch := make(chan int)
	for i := 0; i < b.N; i++ {
		select {
		case <-ch:
		default:
		}
	}
}

// Receive from a closed channel. See CL 216818.
func BenchmarkRecvClosed(b *testing.B) {
	ch := make(chan int)
	close(ch)
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

// Non-blocking receive from a closed channel, the case CL 216818 fixed.
func BenchmarkSelectClosed(b *testing.B) {
	ch := make(chan int)
	close(ch)
	for i := 0; i < b.N; i++ {
		select {
		case <-ch:
		default:
		}
	}
}
