// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Command merge times every merge of package probe at several input counts.
// A cell is one GOMAXPROCS value, one input count n and one merge. Every cell
// gets the same total number of items, split evenly across its n unbuffered
// inputs, with no sleeps. The metric is the practice's bin/practice
// -step=bench: one warm-up run, then -reps timed runs, and the median wall
// time divided by the items delivered. Beside it, CPU time per item: user plus
// system, for the whole process.
//
// Build it first, then run the binary. Do not time `go run`:
//
//	go build -o bin/ ./cmd/merge
//	bin/merge -cpu=1,2,8 -n=2,4,8,16
//	bin/merge -cpu=1 -n=16 -only=select -total=40000 -trace=sel16.out
//
// Each cell prints one line of name=value fields, for a script to read.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/metrics"
	"runtime/trace"
	"slices"
	"strconv"
	"strings"
	"time"

	probe "course/w02/p2-05-probe"
	"course/w02/practice/feed"
	rangeclose "course/w02/practice/step0-range"
)

func main() {
	var (
		cpuFlag    = flag.String("cpu", "1,2,8", "comma-separated GOMAXPROCS values")
		nFlag      = flag.String("n", "2,4,8,16", "comma-separated input counts")
		totalFlag  = flag.Int("total", 400_000, "items per run, split evenly across the inputs")
		repsFlag   = flag.Int("reps", 5, "timed runs per cell; the median is printed")
		rotateFlag = flag.Int("rotate", 0, "rotate the -cpu list, the -n list and the merges by this many places")
		onlyFlag   = flag.String("only", "", "comma-separated merges to run (default all)")
		traceFlag  = flag.String("trace", "", "write an execution trace of one timed run of one cell to this file")
		vFlag      = flag.Bool("v", false, "also print every timed run")
		locksFlag  = flag.Bool("locks", false, "also print the runtime's lock wait per item (see lockWait)")
	)
	flag.Parse()

	fmt.Printf("%s %s/%s  NumCPU=%d  GOMAXPROCS=%d\n", runtime.Version(), runtime.GOOS,
		runtime.GOARCH, runtime.NumCPU(), runtime.GOMAXPROCS(0))

	cpus, err := parseList("-cpu", *cpuFlag)
	if err != nil {
		fail(err)
	}
	ns, err := parseList("-n", *nFlag)
	if err != nil {
		fail(err)
	}
	for _, n := range ns {
		if *totalFlag%n != 0 {
			fail(fmt.Errorf("-total=%d is not a multiple of n=%d", *totalFlag, n))
		}
	}
	var only []string
	if *onlyFlag != "" {
		only = strings.Split(*onlyFlag, ",")
	}
	reps := *repsFlag
	if *traceFlag != "" {
		if len(cpus) != 1 || len(ns) != 1 || len(only) != 1 {
			fail(fmt.Errorf("-trace needs exactly one -cpu, one -n and one -only"))
		}
		reps = 1
	}

	fmt.Printf("total %d items per run, unbuffered, no sleeps; median of %d timed runs after one warm-up; rotate=%d\n",
		*totalFlag, reps, *rotateFlag)
	for _, p := range rotate(cpus, *rotateFlag) {
		runtime.GOMAXPROCS(p)
		for _, n := range rotate(ns, *rotateFlag) {
			for _, m := range rotate(probe.Merges(n), *rotateFlag) {
				if only != nil && !slices.Contains(only, m.Name) {
					continue
				}
				cell(p, n, m, *totalFlag, reps, *traceFlag, *vFlag, *locksFlag)
			}
		}
	}
}

// cell times one merge at one GOMAXPROCS and one input count, and prints its
// line: median, fastest and slowest wall ns/item, median CPU ns/item, the
// ratio of median CPU to median wall, and the check of every run. With locks,
// also the median lock wait per item.
func cell(p, n int, m probe.Merge, total, reps int, traceFile string, verbose, locks bool) {
	per := total / n
	want := map[string]int{}
	names := make([]string, n)
	for k := range n {
		names[k] = "in" + strconv.Itoa(k)
		want[names[k]] = per
	}
	var walls, cpus, waits []float64
	var err error
	for i := range reps + 1 {
		ins := make([]<-chan feed.Item, n)
		for k := range n {
			ins[k] = feed.Source(names[k], per, 0, 0)
		}
		if i > 0 && traceFile != "" {
			startTrace(traceFile)
		}
		w0 := lockWait(locks)
		start, cpu0 := time.Now(), cpuNow()
		got := rangeclose.Collect(m.Run(ins...))
		d, c := time.Since(start), cpuNow()-cpu0
		w := lockWait(locks) - w0
		if i > 0 && traceFile != "" {
			trace.Stop()
		}
		if e := feed.Check(got, want); e != nil && err == nil {
			err = e
		}
		if i == 0 {
			continue
		}
		walls = append(walls, float64(d.Nanoseconds())/float64(total))
		cpus = append(cpus, float64(c.Nanoseconds())/float64(total))
		waits = append(waits, w/float64(total))
		if verbose {
			fmt.Printf("run  cpu=%d n=%d merge=%s rep=%d wall=%.1f cpu-time=%.1f", p, n, m.Name, i, walls[i-1], cpus[i-1])
			if locks {
				fmt.Printf(" lockwait=%.1f", waits[i-1])
			}
			fmt.Println()
		}
	}
	wall, cpu := median(walls), median(cpus)
	extra := ""
	if locks {
		extra = fmt.Sprintf(" lockwait-ns/item=%.1f", median(waits))
	}
	fmt.Printf("cell cpu=%d n=%d merge=%s ns/item=%.1f min=%.1f max=%.1f cpu-ns/item=%.1f cpu/wall=%.2f%s  %s\n",
		p, n, m.Name, wall, slices.Min(walls), slices.Max(walls), cpu, cpu/wall, extra, verdict(err))
}

// lockWait is the runtime's /sync/mutex/wait/total:seconds in ns: time spent
// blocked on a sync.Mutex or a runtime-internal lock, such as a channel's.
// For a runtime lock it counts only waits that went past spinning to sleep,
// and it times one in eight of those and multiplies by eight
// (runtime/lock_spinbit.go lock2, runtime/mprof.go mLockProfile). Nothing in
// these merges takes a sync.Mutex, so here it is the runtime locks. 0 if off.
func lockWait(on bool) float64 {
	if !on {
		return 0
	}
	s := []metrics.Sample{{Name: "/sync/mutex/wait/total:seconds"}}
	metrics.Read(s)
	return s[0].Value.Float64() * 1e9
}

func startTrace(name string) {
	f, err := os.Create(name)
	if err != nil {
		fail(err)
	}
	if err := trace.Start(f); err != nil {
		fail(err)
	}
}

// median is the middle value of xs, as -step=bench takes it: for an even
// count, the upper of the two middle values.
func median(xs []float64) float64 {
	s := slices.Clone(xs)
	slices.Sort(s)
	return s[len(s)/2]
}

// rotate returns xs rotated left by k places, so that successive rounds run
// the cells in different orders.
func rotate[T any](xs []T, k int) []T {
	if len(xs) == 0 {
		return xs
	}
	k %= len(xs)
	return append(slices.Clone(xs[k:]), xs[:k]...)
}

// cpuNow is this process's CPU time so far, or 0 where it cannot be measured.
func cpuNow() time.Duration {
	c, _ := feed.CPUTime()
	return c
}

func verdict(err error) string {
	if err != nil {
		return "WRONG: " + err.Error()
	}
	return "check ok"
}

func parseList(name, s string) ([]int, error) {
	var xs []int
	for f := range strings.SplitSeq(s, ",") {
		x, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil || x < 1 {
			return nil, fmt.Errorf("%s=%q: want a list of positive integers", name, s)
		}
		xs = append(xs, x)
	}
	return xs, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "merge:", err)
	os.Exit(2)
}
