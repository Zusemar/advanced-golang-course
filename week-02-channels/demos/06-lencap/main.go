package main

import (
	"fmt"
	"runtime"
	"time"
)

var data int

func main() {
	// A time.Timer channel has a 1-element buffer, but len and cap
	// are special-cased to report 0 (src/runtime/chan.go:818-836).
	t := time.NewTimer(10 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("timer channel after it fired: len=%d cap=%d\n", len(t.C), cap(t.C))

	// Polling len() looks like waiting, but it establishes no
	// happens-before edge. Run this under -race.
	ch := make(chan struct{}, 1)
	go func() {
		data = 42
		ch <- struct{}{}
	}()
	for len(ch) == 0 {
		runtime.Gosched()
	}
	fmt.Println("read via len() poll:", data)

	// close() racing with len() is a race the detector deliberately ignores
	// (commit 83dfc3b001, CL 135698, "runtime: ignore races between close and len/cap").
	c2 := make(chan int, 4)
	go close(c2)
	for i := 0; i < 1000; i++ {
		_ = len(c2)
		_ = cap(c2)
	}
	fmt.Println("close vs len/cap: no report")
}
