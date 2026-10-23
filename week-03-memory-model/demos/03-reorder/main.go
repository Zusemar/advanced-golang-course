package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

type padded struct {
	v  int32
	av atomic.Int32
	_  [56]byte // separate cache lines; sharing one would only hide the effect
}

var x, y padded
var r1, r2 int32

var seq atomic.Int32 // round number, published by A
var ack atomic.Int32 // round number, acknowledged by B

const rounds = 1_000_000

func run(useAtomic bool) int {
	seq.Store(0)
	ack.Store(0)
	fin := make(chan int)

	go func() { // B
		runtime.LockOSThread()
		for i := int32(1); i <= rounds; i++ {
			for seq.Load() != i {
			}
			if useAtomic {
				y.av.Store(1)
				r2 = x.av.Load()
			} else {
				y.v = 1
				r2 = x.v
			}
			ack.Store(i)
		}
	}()

	go func() { // A, also the referee
		runtime.LockOSThread()
		both := 0
		for i := int32(1); i <= rounds; i++ {
			x.v, y.v, r1, r2 = 0, 0, 0, 0
			x.av.Store(0)
			y.av.Store(0)
			seq.Store(i) // release: B may now enter round i
			if useAtomic {
				x.av.Store(1)
				r1 = y.av.Load()
			} else {
				x.v = 1
				r1 = y.v
			}
			for ack.Load() != i {
			}
			if r1 == 0 && r2 == 0 {
				both++
			}
		}
		fin <- both
	}()
	return <-fin
}

func main() {
	runtime.GOMAXPROCS(2)
	fmt.Printf("rounds=%d each\n\n", rounds)

	plain := run(false)
	fmt.Printf("plain  int32: r1==0 && r2==0 seen %6d times (%.4f%%)\n",
		plain, 100*float64(plain)/float64(rounds))

	at := run(true)
	fmt.Printf("atomic.Int32: r1==0 && r2==0 seen %6d times (%.4f%%)\n",
		at, 100*float64(at)/float64(rounds))

	fmt.Println("\nSequential consistency forbids this outcome. x86-TSO permits it")
	fmt.Println("for plain accesses; sync/atomic restores it.")
}
