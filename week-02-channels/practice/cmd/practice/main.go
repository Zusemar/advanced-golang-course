// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Command practice runs one step of the week 2 session and prints what it
// did: items delivered and checked, wall time, CPU time, goroutines left.
//
// Build it first, then run the binary. Do not time `go run`:
//
//	go build -o bin/ ./cmd/practice
//	bin/practice -step=0             # range drains a closed channel
//	bin/practice -step=2             # the select that spins
//	bin/practice -step=5             # drain to close, nothing leaks
//	bin/practice -step=bench -cpu=1,2,8
//
// -step=bench times step 1 (a goroutine per input) against step 3 (one select
// loop) on the same inputs.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"course/w02/practice/feed"
	rangeclose "course/w02/practice/step0-range"
	goroutines "course/w02/practice/step1-goroutines"
	selectspin "course/w02/practice/step2-select-spin"
	selectnil "course/w02/practice/step3-select-nil"
	earlyexit "course/w02/practice/step4-early-exit"
	drain "course/w02/practice/step5-drain"
)

func main() {
	var (
		stepFlag  = flag.String("step", "0", "step to run: 0 to 5, or bench")
		cpuFlag   = flag.String("cpu", "1,2,8", "bench only: comma-separated GOMAXPROCS values")
		itemsFlag = flag.Int("items", 200_000, "bench only: items per input")
		repsFlag  = flag.Int("reps", 5, "bench only: timed runs per cell; the median is printed")
	)
	flag.Parse()

	fmt.Printf("%s %s/%s  NumCPU=%d  GOMAXPROCS=%d\n", runtime.Version(), runtime.GOOS,
		runtime.GOARCH, runtime.NumCPU(), runtime.GOMAXPROCS(0))

	switch *stepFlag {
	case "0":
		step0()
	case "1":
		step1()
	case "2", "3":
		step23(*stepFlag)
	case "4", "5":
		step45(*stepFlag)
	case "bench":
		cpus, err := parseCPUs(*cpuFlag)
		if err != nil {
			fail(err)
		}
		bench(cpus, *itemsFlag, *repsFlag)
	default:
		fail(fmt.Errorf("-step=%q: want 0 to 5, or bench", *stepFlag))
	}
}

// step0: a source fills a buffer and closes before anyone receives.
func step0() {
	ch := feed.Source("a", 5, 0, 8)
	time.Sleep(20 * time.Millisecond) // long enough for the source to send all 5 and close
	fmt.Printf("source sent 5 items into a buffer of 8, then closed it\n")
	fmt.Printf("before the first receive:  len=%d cap=%d\n", len(ch), cap(ch))
	got := rangeclose.Collect(ch)
	fmt.Printf("range received %d items, then stopped            %s\n", len(got), verdict(feed.Check(got, map[string]int{"a": 5})))
	it, ok := <-ch
	fmt.Printf("one more receive:          v=%+v ok=%v\n", it, ok)
}

// step1: three inputs at three speeds, merged without select.
func step1() {
	want := map[string]int{"fast": 1000, "slow": 50, "slower": 20}
	base := runtime.NumGoroutine()
	start, cpu0 := time.Now(), cpuNow()
	out := goroutines.Merge(
		feed.Source("fast", 1000, 0, 0),
		feed.Source("slow", 50, 2*time.Millisecond, 0),
		feed.Source("slower", 20, 5*time.Millisecond, 0),
	)
	got := rangeclose.Collect(out)
	report(got, want, start, cpu0)
	fmt.Printf("goroutines left behind:  %d\n", settle(base)-base)
}

// step23: one fast input that closes early, one slow input that keeps going.
func step23(step string) {
	want := map[string]int{"fast": 1000, "slow": 50}
	a := feed.Source("fast", 1000, 0, 0)
	b := feed.Source("slow", 50, 2*time.Millisecond, 0)
	var wasted int64
	start, cpu0 := time.Now(), cpuNow()
	var out <-chan feed.Item
	if step == "2" {
		out = selectspin.Merge(a, b, &wasted)
	} else {
		out = selectnil.Merge(a, b, &wasted)
	}
	got := rangeclose.Collect(out)
	report(got, want, start, cpu0)
	fmt.Printf("receives from a closed input:  %d\n", wasted)
}

// step45: take the first 10 items of two 200-item inputs, then stop.
func step45(step string) {
	base := runtime.NumGoroutine()
	a := feed.Source("a", 200, time.Millisecond, 0)
	b := feed.Source("b", 200, time.Millisecond, 0)
	start := time.Now()
	var got []feed.Item
	if step == "4" {
		got = earlyexit.FirstK(10, a, b)
	} else {
		got = drain.FirstK(10, a, b)
	}
	d := time.Since(start)
	fmt.Printf("FirstK returned %d items in %v   %s\n", len(got), d.Round(time.Millisecond), verdict(feed.CheckPrefix(got)))
	fmt.Printf("goroutines left behind:  %d\n", settle(base)-base)
}

func report(got []feed.Item, want map[string]int, start time.Time, cpu0 time.Duration) {
	wall := time.Since(start)
	cpu := "n/a"
	if c := cpuNow(); c >= 0 && cpu0 >= 0 {
		cpu = (c - cpu0).Round(time.Millisecond).String()
	}
	fmt.Printf("delivered %d items   %s\n", len(got), verdict(feed.Check(got, want)))
	fmt.Printf("wall %v   CPU %s\n", wall.Round(time.Millisecond), cpu)
}

// bench times both merges on two inputs of n items each, unbuffered, no sleeps.
func bench(cpus []int, n, reps int) {
	type merge struct {
		name string
		run  func(a, b <-chan feed.Item) <-chan feed.Item
	}
	merges := []merge{
		{"1-goroutines", func(a, b <-chan feed.Item) <-chan feed.Item { return goroutines.Merge(a, b) }},
		{"3-select-nil", func(a, b <-chan feed.Item) <-chan feed.Item {
			var w int64
			return selectnil.Merge(a, b, &w)
		}},
	}
	fmt.Printf("two inputs x %d items, unbuffered; median of %d timed runs after one warm-up\n\n", n, reps)
	fmt.Printf("%-4s %-14s %12s %10s  %s\n", "cpu", "merge", "median", "ns/item", "check")
	want := map[string]int{"a": n, "b": n}
	for _, p := range cpus {
		runtime.GOMAXPROCS(p)
		for _, m := range merges {
			var ds []time.Duration
			var err error
			for i := range reps + 1 {
				a, b := feed.Source("a", n, 0, 0), feed.Source("b", n, 0, 0)
				start := time.Now()
				got := rangeclose.Collect(m.run(a, b))
				d := time.Since(start)
				err = feed.Check(got, want)
				if i > 0 {
					ds = append(ds, d)
				}
			}
			slices.Sort(ds)
			med := ds[len(ds)/2]
			fmt.Printf("%-4d %-14s %12v %10.0f  %s\n", p, m.name, med.Round(100*time.Microsecond),
				float64(med.Nanoseconds())/float64(2*n), verdict(err))
		}
	}
}

func cpuNow() time.Duration {
	c, ok := feed.CPUTime()
	if !ok {
		return -1
	}
	return c
}

// settle waits up to half a second for the goroutine count to fall back to
// base, and returns the last count it saw.
func settle(base int) int {
	n := runtime.NumGoroutine()
	for deadline := time.Now().Add(500 * time.Millisecond); n > base && time.Now().Before(deadline); {
		time.Sleep(10 * time.Millisecond)
		n = runtime.NumGoroutine()
	}
	return n
}

func verdict(err error) string {
	if err != nil {
		return "WRONG: " + err.Error()
	}
	return "check ok"
}

func parseCPUs(s string) ([]int, error) {
	var cpus []int
	for f := range strings.SplitSeq(s, ",") {
		p, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil || p < 1 {
			return nil, fmt.Errorf("-cpu=%q: want a list of positive integers", s)
		}
		cpus = append(cpus, p)
	}
	return cpus, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "practice:", err)
	os.Exit(2)
}
