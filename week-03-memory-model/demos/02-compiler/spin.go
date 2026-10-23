package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

var done bool
var flag atomic.Bool
var shared int

// A. The memory model's "busy waiting for a value". The doc warns:
// "there is no guarantee that the write to done will ever be observed by
// main... The loop in main is not guaranteed to finish."
func SpinPlain() {
	for !done {
	}
}

// B. Same loop through sync/atomic.
func SpinAtomic() {
	for !flag.Load() {
	}
}

// C. The one rewrite the memory model explicitly PERMITS
// ("Incorrect compilation", final example): hoist the load of a shared
// variable out of a loop, because the extra read cannot affect any existing
// concurrent access.
func SumShared(m int) int {
	n := 0
	for i := 0; i < m; i++ {
		n += shared
	}
	return n
}

// SpinPlain terminates on gc/amd64 today because the compiler reloads `done`
// on every iteration (see run.sh). Nothing guarantees that; the same source
// under a compiler that hoists the load would hang.
func main() {
	go func() { done = true }()
	spun := make(chan struct{})
	go func() { SpinPlain(); close(spun) }()
	select {
	case <-spun:
		fmt.Println("plain-bool spin loop: terminated")
	case <-time.After(2 * time.Second):
		fmt.Println("plain-bool spin loop: STILL SPINNING after 2s")
	}

	go func() { flag.Store(true) }()
	spun2 := make(chan struct{})
	go func() { SpinAtomic(); close(spun2) }()
	select {
	case <-spun2:
		fmt.Println("atomic.Bool spin loop: terminated")
	case <-time.After(2 * time.Second):
		fmt.Println("atomic.Bool spin loop: STILL SPINNING after 2s")
	}
}
