// Demo 1: Reading GODEBUG=schedtrace output.
//
// Build first, then run the binary. `go run` under GODEBUG=schedtrace also
// traces the go command and the compiler, a few hundred extra lines.
//
//	env -u GOROOT go build -o /tmp/bin/01-schedtrace ./01-schedtrace
//	GOMAXPROCS=2 GODEBUG=schedtrace=1000 /tmp/bin/01-schedtrace
//	GOMAXPROCS=2 GODEBUG=schedtrace=1000,scheddetail=1 /tmp/bin/01-schedtrace
//
// The program deliberately creates more runnable goroutines than Ps, so the
// run queues are visibly non-empty in the trace.
package main

import (
	"fmt"
	"runtime"
	"time"
)

func burn(d time.Duration) {
	end := time.Now().Add(d)
	x := 0
	for time.Now().Before(end) {
		x++
	}
	_ = x
}

func main() {
	fmt.Printf("GOMAXPROCS=%d NumCPU=%d\n", runtime.GOMAXPROCS(0), runtime.NumCPU())

	// 20 CPU-bound goroutines on (probably) 2 Ps. The 256-slot local run queues
	// never fill; they hold a handful each. The global queue is the one that
	// grows, because a goroutine sysmon preempts goes to the global queue
	// (goschedImpl -> globrunqput), not back to its P's local queue.
	for i := 0; i < 20; i++ {
		go burn(4 * time.Second)
	}
	time.Sleep(4500 * time.Millisecond)
}
