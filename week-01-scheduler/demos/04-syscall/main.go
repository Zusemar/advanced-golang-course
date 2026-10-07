// Demo 4: Syscall handoff, made visible with runtime/metrics (Go 1.26+).
//
// A goroutine that blocks in a syscall the netpoller cannot handle keeps its
// M. sysmon notices after ~20us and takes the P away, handing it to a fresh
// thread. So N goroutines blocked in blocking syscalls cost N OS threads,
// while N goroutines blocked on the network cost approximately zero.
//
// Run at GOMAXPROCS=2, the default on the 2-CPU box the lecture's transcript
// was captured on. At a larger GOMAXPROCS the network line grows with it:
//
//	GOMAXPROCS=2 go run ./04-syscall
package main

import (
	"fmt"
	"net"
	"os"
	"runtime/metrics"
	"strings"
	"sync"
	"time"
)

var names = []string{
	"/sched/gomaxprocs:threads",
	"/sched/threads/total:threads",
	"/sched/goroutines:goroutines",
	"/sched/goroutines/running:goroutines",
	"/sched/goroutines/runnable:goroutines",
	"/sched/goroutines/waiting:goroutines",
	"/sched/goroutines/not-in-go:goroutines",
}

func snapshot(label string) {
	s := make([]metrics.Sample, len(names))
	for i, n := range names {
		s[i].Name = n
	}
	metrics.Read(s)
	fmt.Printf("%-26s", label)
	for i := range s {
		short := strings.TrimPrefix(names[i], "/sched/")
		short = short[:strings.IndexByte(short, ':')]
		fmt.Printf("  %s=%d", short, s[i].Value.Uint64())
	}
	fmt.Println()
}

func main() {
	snapshot("baseline")

	// (a) 200 goroutines blocked on the NETWORK. The netpoller parks them on
	// epoll; no thread is consumed.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("listen:", err)
		os.Exit(1)
	}
	defer ln.Close()
	var conns []net.Conn
	var connsMu sync.Mutex // the accept goroutine and main both append
	var wgNet sync.WaitGroup
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			connsMu.Lock()
			conns = append(conns, c)
			connsMu.Unlock()
		}
	}()
	for range 200 {
		c, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			break
		}
		connsMu.Lock()
		conns = append(conns, c)
		connsMu.Unlock()
		wgNet.Go(func() {
			buf := make([]byte, 1)
			c.SetReadDeadline(time.Now().Add(3 * time.Second))
			c.Read(buf) // parks on the netpoller
		})
	}
	time.Sleep(300 * time.Millisecond)
	snapshot("200 blocked on network")

	// (b) 200 goroutines blocked in a BLOCKING syscall the poller cannot
	// take: nanosleep(2), in blocking.go, has no file descriptor to poll, so
	// each one keeps its M in the kernel. A read(2) on a regular file would
	// do the same (epoll_ctl refuses regular files with EPERM), but this
	// demo reads no file.
	var wgSys sync.WaitGroup
	for range 200 {
		wgSys.Go(func() { blockingSyscall(2 * time.Second) })
	}
	time.Sleep(500 * time.Millisecond)
	snapshot("200 in blocking syscalls")

	wgSys.Wait()
	time.Sleep(200 * time.Millisecond)
	snapshot("after they return")
}
