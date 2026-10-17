// L2-02 probe. S senders send unique values into one buffered channel at once; one
// receiver takes them all. With the stock runtime every value arrives exactly once.
//
// Built against runtime/chan.go patched by nolock.diff (the ring write in chansend moved
// out from under c.lock, chan.go:236-249 at go1.27.1), the senders no longer take turns
// on sendx, qcount and the slots. The probe counts, from user code only:
//
//	lost     sent, never received
//	dup      received more than once
//	phantom  received, never sent (0 is never sent: it is what a cleared slot holds)
//	over     receives after which len(ch) > cap(ch)
//	len0     len(ch) of the new channel before the first send (must be 0)
//
// Values are s*per+i+1, so every one is > 0 and names its sender and its place.
// Built with -tags l202count against chan.go patched by guard.diff, it also prints the
// runtime's own counters: ring writes, writes after which qcount > dataqsiz, writes that
// found sendx outside the buffer, and writes into a slot that still held a value.
// See L2-02.md, «Проверить самому».
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	senders = flag.Int("senders", 16, "goroutines sending at once")
	per     = flag.Int("per", 100000, "values per sender")
	capa    = flag.Int("cap", 4, "channel buffer size")
	trials  = flag.Int("trials", 5, "trials, each on a fresh channel")
)

var progress atomic.Int64

type result struct {
	received, lost, dup, phantom, over int64
	maxlen, len0                       int
}

func trial() result {
	S, P := *senders, *per
	total := S * P
	ch := make(chan int, *capa)
	seen := make([]uint8, total+1)
	var r result
	r.len0 = len(ch) // 0 for every new channel, unless something wrote into its memory before make
	done := make(chan struct{})
	go func() { // the receiver
		for v := range ch {
			r.received++
			progress.Add(1)
			if v <= 0 || v > total {
				r.phantom++
			} else if seen[v] < 255 {
				seen[v]++
			}
			if l := len(ch); l > cap(ch) {
				r.over++
				if l > r.maxlen {
					r.maxlen = l
				}
			} else if l > r.maxlen {
				r.maxlen = l
			}
		}
		close(done)
	}()
	var wg sync.WaitGroup
	start := make(chan struct{})
	for s := range S {
		wg.Add(1)
		go func(s int) {
			defer wg.Done()
			<-start
			for i := range P {
				ch <- s*P + i + 1
			}
		}(s)
	}
	close(start)
	wg.Wait()
	close(ch)
	<-done
	for v := 1; v <= total; v++ {
		switch {
		case seen[v] == 0:
			r.lost++
		case seen[v] > 1:
			r.dup++
		}
	}
	return r
}

func main() {
	flag.Parse()
	go func() { // watchdog: a lost wake-up or a broken ring may leave the receiver stuck
		last := int64(-1)
		for {
			time.Sleep(5 * time.Second)
			p := progress.Load()
			if p == last {
				fmt.Printf("STUCK after %d receives\n", p)
				os.Exit(3)
			}
			last = p
		}
	}()
	fmt.Printf("# senders=%d per=%d cap=%d trials=%d GOMAXPROCS=%d NumCPU=%d counting=%v %s\n",
		*senders, *per, *capa, *trials, runtime.GOMAXPROCS(0), runtime.NumCPU(), counting, runtime.Version())
	var bad int
	for t := 1; t <= *trials; t++ {
		w0, o0, b0, v0 := ringCounts()
		t0 := time.Now()
		r := trial()
		el := time.Since(t0)
		w1, o1, b1, v1 := ringCounts()
		sent := int64(*senders * *per)
		fmt.Printf("trial=%d sent=%d received=%d lost=%d dup=%d phantom=%d over=%d maxlen=%d len0=%d ms=%d",
			t, sent, r.received, r.lost, r.dup, r.phantom, r.over, r.maxlen, r.len0, el.Milliseconds())
		if counting {
			fmt.Printf(" ringwrites=%d overfull=%d badindex=%d overwrite=%d", w1-w0, o1-o0, b1-b0, v1-v0)
		}
		fmt.Println()
		if r.lost+r.dup+r.phantom+r.over != 0 || r.len0 != 0 || r.received != sent {
			bad++
		}
	}
	fmt.Printf("trials_bad=%d of %d\n", bad, *trials)
}
