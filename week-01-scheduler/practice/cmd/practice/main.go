// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Command practice runs the week 1 executor steps and prints how long each
// took, its speedup over step 0, and what it did to the scheduler.
//
// Build it first, then run the binary. Do not time `go run`:
//
//	go build -o bin/ ./cmd/practice
//	bin/practice -step=0 -cpu=1                  # the baseline
//	bin/practice -step=all -cpu=1,2,8            # the speedup table
//	bin/practice -step=3 -cpu=8 -trace=step3.out # one run, traced
//	go tool trace -http=127.0.0.1:6060 step3.out
//	go tool trace -pprof=sched step4.out > sched.prof && go tool pprof -top sched.prof
//
// Run `go tool trace` from inside this module. Outside the repository `go` may
// be an older toolchain, and it cannot read a go1.27 trace.
//
// -cpu sets GOMAXPROCS for each row, as `go test -cpu` does. The executor's
// bound n is the same number unless -n says otherwise.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"slices"
	"strconv"
	"strings"
	"time"

	"course/w01/practice"
	"course/w01/practice/work"
)

func main() {
	var (
		stepFlag  = flag.String("step", "all", "step to run: 0 to 5, or all")
		cpuFlag   = flag.String("cpu", "1,2,8", "comma-separated GOMAXPROCS values, one row each")
		nFlag     = flag.Int("n", 0, "executor bound; 0 means the row's -cpu value")
		tasksFlag = flag.Int("tasks", 1000, "number of tasks")
		shapeFlag = flag.String("shape", "skewed", "task cost shape: skewed (task i costs i+1) or even")
		repsFlag  = flag.Int("reps", 5, "timed runs per cell; the median is printed")
		traceFlag = flag.String("trace", "", "write an execution trace of one run to this file")
	)
	flag.Parse()

	steps, err := pickSteps(*stepFlag)
	if err != nil {
		fail(err)
	}
	cpus, err := parseCPUs(*cpuFlag)
	if err != nil {
		fail(err)
	}
	if _, err := work.New(1, *shapeFlag); err != nil {
		fail(err)
	}

	fmt.Printf("%s %s/%s  NumCPU=%d  tasks=%d shape=%s\n",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(),
		*tasksFlag, *shapeFlag)
	for _, p := range cpus {
		if p > runtime.NumCPU() {
			fmt.Printf("note: -cpu=%d is more than NumCPU=%d; nothing past %d can run in parallel\n",
				p, runtime.NumCPU(), runtime.NumCPU())
		}
	}

	if *traceFlag != "" {
		if len(steps) != 1 || len(cpus) != 1 {
			fail(fmt.Errorf("-trace needs exactly one -step and one -cpu value"))
		}
		p := cpus[0]
		runtime.GOMAXPROCS(p)
		n := bound(*nFlag, p)
		d, l := traced(*traceFlag, steps[0], n, *tasksFlag, *shapeFlag)
		fmt.Printf("%s  cpu=%d n=%d  %v  peak running %d, peak goroutines %d  %s\n",
			steps[0].Name, p, n, d.Round(100*time.Microsecond),
			l.PeakRunning(), l.PeakGoroutines(), verdict(l))
		fmt.Printf("trace written to %s\n", *traceFlag)
		return
	}

	fmt.Printf("median of %d timed runs after one warm-up; running and goroutines are peaks from the last run\n", *repsFlag)
	fmt.Printf("\n%-4s %-4s %-14s %10s %8s %8s %11s  %s\n",
		"cpu", "n", "step", "median", "speedup", "running", "goroutines", "check")
	for _, p := range cpus {
		runtime.GOMAXPROCS(p)
		n := bound(*nFlag, p)
		// The baseline is re-measured at every -cpu value, so the speedup
		// column never divides by a number taken under different conditions.
		// Step 0's own row is that measurement, so it reads 1.00x by
		// construction; any other step at -cpu=1 shows the run-to-run noise.
		base, baseLoad := measure(practice.Steps[0], n, *tasksFlag, *shapeFlag, *repsFlag)
		for _, s := range steps {
			d, l := base, baseLoad
			if s.Name != practice.Steps[0].Name {
				d, l = measure(s, n, *tasksFlag, *shapeFlag, *repsFlag)
			}
			fmt.Printf("%-4d %-4d %-14s %10v %7.2fx %8d %11d  %s\n",
				p, n, s.Name, d.Round(100*time.Microsecond), float64(base)/float64(d),
				l.PeakRunning(), l.PeakGoroutines(), verdict(l))
		}
	}
}

// measure runs s once untimed, then reps times timed, and returns the median
// and the last run's load. The peaks are from that one run, not a maximum.
func measure(s practice.Step, n, tasks int, shape string, reps int) (time.Duration, *work.Load) {
	var l *work.Load
	ds := make([]time.Duration, 0, reps)
	for i := range reps + 1 {
		l, _ = work.New(tasks, shape)
		start := time.Now()
		s.Run(n, l.Tasks())
		d := time.Since(start)
		if i > 0 {
			ds = append(ds, d)
		}
	}
	slices.Sort(ds) // an unsorted "median" is a random pick
	return ds[len(ds)/2], l
}

// traced runs s once under the execution tracer. Each task is wrapped in a
// trace region named "task", so the goroutine view shows where tasks ran.
func traced(path string, s practice.Step, n, tasks int, shape string) (time.Duration, *work.Load) {
	f, err := os.Create(path)
	if err != nil {
		fail(err)
	}
	defer f.Close()

	l, _ := work.New(tasks, shape)
	ctx := context.Background()
	wrapped := make([]func(), len(l.Tasks()))
	for i, t := range l.Tasks() {
		wrapped[i] = func() { trace.WithRegion(ctx, "task", t) }
	}
	if err := trace.Start(f); err != nil {
		fail(err)
	}
	start := time.Now()
	trace.WithRegion(ctx, s.Name, func() { s.Run(n, wrapped) })
	d := time.Since(start)
	trace.Stop()
	return d, l
}

func verdict(l *work.Load) string {
	if err := l.Check(); err != nil {
		return "WRONG: " + err.Error()
	}
	return "ok"
}

func bound(n, p int) int {
	if n > 0 {
		return n
	}
	return p
}

func pickSteps(s string) ([]practice.Step, error) {
	if s == "all" {
		return practice.Steps, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil || i < 0 || i >= len(practice.Steps) {
		return nil, fmt.Errorf("-step=%q: want 0 to %d, or all", s, len(practice.Steps)-1)
	}
	return practice.Steps[i : i+1], nil
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
