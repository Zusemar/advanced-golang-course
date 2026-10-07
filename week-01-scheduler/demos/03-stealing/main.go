// Demo 3: The 256-slot local run queue, and what overflows out of it.
//
// A goroutine that spawns work puts each new G on ITS OWN P's local run queue
// (runqput). That queue is a fixed 256-entry ring. When it is full, runqputslow
// moves half of it plus the new G to the global queue under sched.lock.
//
// So a spawn loop of 1000 goroutines from a single goroutine is not "1000 Gs
// evenly distributed". It is: some in one local queue, most in the global
// queue, and whatever other Ps manage to steal.
//
// Build first, then run the binary. `go run` under GODEBUG=schedtrace also
// traces the go command and the compiler, a few hundred extra lines.
//
//	env -u GOROOT go build -o /tmp/bin/03-stealing ./03-stealing
//	GOMAXPROCS=1 GODEBUG=schedtrace=100 /tmp/bin/03-stealing
//
// Watch `runqueue=` (the GLOBAL queue) and the per-P bracket `[ n ]`.
// The local number never exceeds 256.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Printf("GOMAXPROCS=%d\n", runtime.GOMAXPROCS(0))

	var wg sync.WaitGroup
	// Spawn all 1000 from ONE goroutine, so they all target one local queue.
	for range 1000 {
		wg.Go(func() {
			// Just enough work that they do not all drain instantly.
			t := time.Now()
			for time.Since(t) < 2*time.Millisecond {
			}
		})
	}
	time.Sleep(400 * time.Millisecond) // let schedtrace print mid-flight
	wg.Wait()
	fmt.Println("done")
}
