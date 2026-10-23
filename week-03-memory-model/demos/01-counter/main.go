package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	goroutines = 8
	perG       = 100_000
	want       = goroutines * perG
)

func broken() int64 {
	var n int64
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perG {
				n++
			}
		})
	}
	wg.Wait()
	return n
}

func withMutex() int64 {
	var n int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perG {
				mu.Lock()
				n++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	return n
}

func withAtomic() int64 {
	var n atomic.Int64
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perG {
				n.Add(1)
			}
		})
	}
	wg.Wait()
	return n.Load()
}

func withOwner() int64 {
	// The counter is owned by exactly one goroutine; everyone else sends.
	ch := make(chan int64, 1024)
	done := make(chan int64)
	go func() {
		var n int64
		for d := range ch {
			n += d
		}
		done <- n
	}()
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perG {
				ch <- 1
			}
		})
	}
	wg.Wait()
	close(ch)
	return <-done
}

func main() {
	fmt.Printf("want %d\n", want)
	for _, c := range []struct {
		name string
		f    func() int64
	}{
		{"unsynchronized", broken},
		{"sync.Mutex", withMutex},
		{"atomic.Int64", withAtomic},
		{"channel-owned", withOwner},
	} {
		got := c.f()
		fmt.Printf("%-16s got %8d  lost %6d\n", c.name, got, want-int64(got))
	}
}
