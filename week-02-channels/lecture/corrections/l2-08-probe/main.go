// L2-08 probe. One goroutine parks in a select over n channels. Then n other goroutines
// wake it at the same moment, each through its own channel. Only one may win.
//
// With the stock runtime it checks that every select returns one case and every value
// arrives exactly once. Built with -tags l208count against runtime/chan.go patched by
// count.diff, runtime.SelectRaceCounts reports how many wakers won and lost the CAS on
// g.selectDone in waitq.dequeue (chan.go:896-901 at go1.27.1). With nocas.diff the CAS
// is gone and every waker "wins". See L2-08.md, «Проверить самому».
//
// Each round: the selector announces the round and enters the select; a coordinator waits
// -settle (so the select is parked) and opens the gate; the n wakers, spinning on the gate,
// send (-dir=recv) or receive (-dir=send) on their own channel. The selector reads the
// counters as soon as the select returns, then completes the other n-1 operations so every
// waker is released. Every value is checked against the round number.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

const maxN = 16

var (
	n      = flag.Int("n", 16, "live channels in the select, 2..16")
	rounds = flag.Int("rounds", 20000, "rounds")
	dir    = flag.String("dir", "recv", "recv: the select receives and n senders wake it; send: the select sends and n receivers wake it")
	settle = flag.Duration("settle", 20*time.Microsecond, "wait after the selector enters the select before the wakers go")
	gap    = flag.Duration("gap", 0, "waker i goes i*gap after the gate opens (0: all at once)")
	gosch  = flag.Bool("gosched", false, "a waker calls runtime.Gosched right after its operation, so the woken select, in runnext, gets the P at once")
)

var (
	ready    atomic.Int64 // the selector is entering the select of this round
	gate     atomic.Int64 // the wakers of this round may go
	progress atomic.Int64
	bad      atomic.Int64
	yield    bool
)

func wait(a *atomic.Int64, r int64) {
	for i := 0; a.Load() < r; i++ {
		if yield && i&63 == 63 {
			runtime.Gosched()
		}
	}
}

func spin(d time.Duration) {
	if d <= 0 {
		return
	}
	t := time.Now()
	for time.Since(t) < d {
	}
}

func main() {
	flag.Parse()
	if *n < 2 || *n > maxN || (*dir != "recv" && *dir != "send") {
		fmt.Fprintln(os.Stderr, "bad flags")
		os.Exit(2)
	}
	yield = runtime.GOMAXPROCS(0) < *n+2
	var cs [maxN]chan int
	for i := 0; i < *n; i++ {
		cs[i] = make(chan int)
	}
	R := int64(*rounds)

	go func() { // coordinator
		for r := int64(1); r <= R; r++ {
			wait(&ready, r)
			spin(*settle)
			gate.Store(r)
		}
	}()
	for i := 0; i < *n; i++ { // wakers
		go func(i int) {
			for r := int64(1); r <= R; r++ {
				wait(&gate, r)
				spin(time.Duration(i) * *gap)
				if *dir == "recv" {
					cs[i] <- int(r)
				} else if v := <-cs[i]; v != int(r) {
					bad.Add(1)
				}
				if *gosch {
					runtime.Gosched()
				}
			}
		}(i)
	}
	go func() { // watchdog
		last := int64(-1)
		for {
			time.Sleep(3 * time.Second)
			p := progress.Load()
			if p == last {
				fmt.Printf("STUCK at round %d\n", p+1)
				os.Exit(3)
			}
			last = p
		}
	}()

	var hist [maxN]int64 // rounds by number of lost CAS
	var parked, won, lost, wonTotal int64
	w0, l0 := raceCounts()
	start := time.Now()
	for r := int64(1); r <= R; r++ {
		ready.Store(r)
		var k int
		if *dir == "recv" {
			var v int
			k, v = selRecv(&cs)
			if v != int(r) {
				bad.Add(1)
			}
		} else {
			k = selSend(&cs, int(r))
		}
		w1, l1 := raceCounts()
		dw, dl := int64(w1-w0), int64(l1-l0)
		w0, l0 = w1, l1
		wonTotal += dw
		if dw > 0 {
			parked++
		}
		if dw > 1 {
			bad.Add(1) // two winners for one select
		}
		won += dw
		lost += dl
		if dl < maxN {
			hist[dl]++
		}
		for j := 0; j < *n; j++ {
			if j == k {
				continue
			}
			if *dir == "recv" {
				if v := <-cs[j]; v != int(r) {
					bad.Add(1)
				}
			} else {
				cs[j] <- int(r)
			}
		}
		progress.Store(r)
	}
	el := time.Since(start)
	var hs []string
	for i := 0; i < *n; i++ {
		hs = append(hs, fmt.Sprint(hist[i]))
	}
	withLoss := R - hist[0]
	fmt.Printf("dir=%s n=%d gap=%v settle=%v gosched=%v rounds=%d gomaxprocs=%d ncpu=%d yield=%v counting=%v | "+
		"parked=%d won=%d lost=%d rounds_with_loss=%d (%.1f%% of parked) lost_per_parked=%.2f hist_lost=[%s] bad=%d elapsed=%v\n",
		*dir, *n, *gap, *settle, *gosch, R, runtime.GOMAXPROCS(0), runtime.NumCPU(), yield, counting,
		parked, won, lost, withLoss, 100*float64(withLoss)/float64(max(parked, 1)),
		float64(lost)/float64(max(parked, 1)), strings.Join(hs, " "), bad.Load(), el.Round(time.Millisecond))
}

func selRecv(c *[maxN]chan int) (int, int) {
	select {
	case v := <-c[0]:
		return 0, v
	case v := <-c[1]:
		return 1, v
	case v := <-c[2]:
		return 2, v
	case v := <-c[3]:
		return 3, v
	case v := <-c[4]:
		return 4, v
	case v := <-c[5]:
		return 5, v
	case v := <-c[6]:
		return 6, v
	case v := <-c[7]:
		return 7, v
	case v := <-c[8]:
		return 8, v
	case v := <-c[9]:
		return 9, v
	case v := <-c[10]:
		return 10, v
	case v := <-c[11]:
		return 11, v
	case v := <-c[12]:
		return 12, v
	case v := <-c[13]:
		return 13, v
	case v := <-c[14]:
		return 14, v
	case v := <-c[15]:
		return 15, v
	}
}

func selSend(c *[maxN]chan int, r int) int {
	select {
	case c[0] <- r:
		return 0
	case c[1] <- r:
		return 1
	case c[2] <- r:
		return 2
	case c[3] <- r:
		return 3
	case c[4] <- r:
		return 4
	case c[5] <- r:
		return 5
	case c[6] <- r:
		return 6
	case c[7] <- r:
		return 7
	case c[8] <- r:
		return 8
	case c[9] <- r:
		return 9
	case c[10] <- r:
		return 10
	case c[11] <- r:
		return 11
	case c[12] <- r:
		return 12
	case c[13] <- r:
		return 13
	case c[14] <- r:
		return 14
	case c[15] <- r:
		return 15
	}
}
