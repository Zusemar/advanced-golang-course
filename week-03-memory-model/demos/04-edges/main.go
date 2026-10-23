package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Every case has the same shape:
//
//	writer: payload = 1 ; RELEASE
//	reader: ACQUIRE      ; observe payload
//
// -race adjudicates: if RELEASE is synchronized before ACQUIRE it says nothing.
var payload, sink int

func reset() { payload, sink = 0, 0 }

//  1. Goroutine creation: "The go statement that starts a new goroutine is
//     synchronized before the start of the goroutine's execution."
func goCreate(done chan struct{}) {
	reset()
	payload = 1
	go func() { sink = payload; close(done) }()
	<-done
}

//  2. Goroutine destruction: "The exit of a goroutine is not guaranteed to be
//     synchronized before any event in the program."  NO EDGE.
func goExit(done chan struct{}) {
	reset()
	go func() { payload = 1 }()
	time.Sleep(time.Millisecond)
	sink = payload // racy: nothing orders the goroutine's exit against this
	close(done)
}

//  3. Channel send: "A send on a channel is synchronized before the completion
//     of the corresponding receive from that channel."
func chanSendRecv(done chan struct{}) {
	reset()
	c := make(chan int)
	go func() { payload = 1; c <- 0 }()
	<-c
	sink = payload
	close(done)
}

//  4. Unbuffered receive: "A receive from an unbuffered channel is synchronized
//     before the completion of the corresponding send on that channel." This is
//     the memory model's own hello-world example.
func chanRecvSendUnbuffered(done chan struct{}) {
	reset()
	c := make(chan int)
	go func() { payload = 1; <-c }()
	c <- 0
	sink = payload
	close(done)
}

//  5. Unbuffered-style recv->send, but with capacity 1. The memory model says
//     explicitly: "If the channel were buffered (e.g., c = make(chan int, 1))
//     then the program would not be guaranteed to print hello, world."  NO EDGE.
func chanRecvSendBuffered(done chan struct{}) {
	reset()
	c := make(chan int, 1)
	go func() { payload = 1; <-c }()
	c <- 0
	sink = payload
	close(done)
}

//  6. Close: "The closing of a channel is synchronized before a receive that
//     returns a zero value because the channel is closed."
func chanClose(done chan struct{}) {
	reset()
	c := make(chan int)
	go func() { payload = 1; close(c) }()
	<-c
	sink = payload
	close(done)
}

//  7. Mutex: "For any sync.Mutex or sync.RWMutex variable l and n < m, call n of
//     l.Unlock() is synchronized before call m of l.Lock() returns."
func mutexEdge(done chan struct{}) {
	reset()
	var mu sync.Mutex
	mu.Lock()
	go func() { payload = 1; mu.Unlock() }()
	mu.Lock() // returns only after the goroutine's Unlock
	sink = payload
	mu.Unlock()
	close(done)
}

//  8. Once: "The completion of a single call of f() from once.Do(f) is
//     synchronized before the return of any call of once.Do(f)."
func onceEdge(done chan struct{}) {
	reset()
	var once sync.Once
	var wg sync.WaitGroup
	seen := make([]int, 2)
	for i := range 2 {
		wg.Go(func() {
			once.Do(func() { payload = 1 })
			seen[i] = payload // distinct slots: only payload is under test
		})
	}
	wg.Wait()
	sink = seen[0] + seen[1]
	close(done)
}

//  9. Atomics: "If the effect of an atomic operation A is observed by atomic
//     operation B, then A is synchronized before B."
func atomicEdge(done chan struct{}) {
	reset()
	var flag atomic.Bool
	go func() { payload = 1; flag.Store(true) }()
	for !flag.Load() {
	}
	sink = payload
	close(done)
}

// 10. Plain bool flag. NO EDGE — the memory model's "busy waiting for a value".
func plainFlag(done chan struct{}) {
	reset()
	var flag bool
	go func() { payload = 1; flag = true }()
	for !flag {
	}
	sink = payload
	close(done)
}

//  11. WaitGroup: the goroutine started by wg.Go finishes before wg.Wait
//     returns, so the write is synchronized before the read.
func waitGroupEdge(done chan struct{}) {
	reset()
	var wg sync.WaitGroup
	wg.Go(func() { payload = 1 })
	wg.Wait()
	sink = payload
	close(done)
}

type edgeCase struct {
	id         string
	name       string
	f          func(chan struct{})
	expectRace bool
}

var cases = []edgeCase{
	{"01", "go statement", goCreate, false},
	{"02", "goroutine exit", goExit, true},
	{"03", "chan send -> recv", chanSendRecv, false},
	{"04", "unbuffered recv -> send", chanRecvSendUnbuffered, false},
	{"05", "cap-1 recv -> send", chanRecvSendBuffered, true},
	{"06", "close -> recv", chanClose, false},
	{"07", "Unlock -> Lock", mutexEdge, false},
	{"08", "sync.Once", onceEdge, false},
	{"09", "atomic Store -> Load", atomicEdge, false},
	{"10", "plain bool flag", plainFlag, true},
	{"11", "WaitGroup.Go / Wait", waitGroupEdge, false},
}

// Run one case per process, so that one -race report cannot mask another:
//
//	go build -race -o /tmp/bin/04-edges ./04-edges
//	for n in 01 02 03 04 05 06 07 08 09 10 11; do /tmp/bin/04-edges $n; done
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: 04-edges NN   (NN = 01 .. 11)")
		os.Exit(2)
	}
	for _, c := range cases {
		if c.id != os.Args[1] {
			continue
		}
		if c.expectRace {
			fmt.Printf("=== %s %s  <-- expect race\n", c.id, c.name)
		} else {
			fmt.Printf("=== %s %s\n", c.id, c.name)
		}
		c.f(make(chan struct{}))
		return
	}
	fmt.Fprintf(os.Stderr, "no such case: %s\n", os.Args[1])
	os.Exit(2)
}
