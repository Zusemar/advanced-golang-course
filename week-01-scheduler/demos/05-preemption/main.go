// Demo 5: Asynchronous preemption, and what the world looked like before Go 1.14.
//
// A goroutine running a loop with no function calls, no allocation and no
// channel operations offers the runtime no cooperative safe-point. Before
// Go 1.14 that goroutine owned its P until it finished.
//
// Build first, then run the binary. `go run` under GODEBUG also applies the
// setting to the go command and the compiler.
//
//	env -u GOROOT go build -o /tmp/bin/05-preemption ./05-preemption
//	GOMAXPROCS=1 /tmp/bin/05-preemption                           # ~15ms, then exits
//	GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 /tmp/bin/05-preemption # hangs; Ctrl-C it
//
// The ~15ms is not arbitrary. sysmon preempts a goroutine that has held its P
// for one time slice, and that slice is
//
//	const forcePreemptNS = 10 * 1000 * 1000 // 10ms
//
// in src/runtime/proc.go. But sysmon only looks every so often: it notices the
// hog on one wakeup, starts the 10ms clock then, and preempts on the first
// wakeup after the 10ms are up. Its sleep backs off from 20us towards 10ms
// while nothing needs retaking, so the observed figure is forcePreemptNS plus
// sysmon's lag, not 10ms flat. Async preemption is HOW it is stopped;
// sysmon's forcePreemptNS is WHEN.
//
// The hang under asyncpreemptoff=1 is golang/go#10958, filed May 2015 and
// fixed February 2020.
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

var sink int

func main() {
	fmt.Printf("GOMAXPROCS=%d  GODEBUG=%q\n", runtime.GOMAXPROCS(0), os.Getenv("GODEBUG"))

	// The hog: a tight loop the compiler cannot insert a safe-point into.
	go func() {
		for i := 0; i < 1<<62; i++ {
			sink += i
		}
	}()

	// Hand the P to the hog. With GOMAXPROCS=1 there is exactly one P, so
	// main cannot run again until something takes it back from the hog.
	start := time.Now()
	runtime.Gosched()

	fmt.Printf("main resumed after %v: the hog was preempted\n",
		time.Since(start).Round(100*time.Microsecond))
}
