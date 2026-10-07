// Demo 6: Where GOMAXPROCS comes from (Go 1.25+).
//
// On Linux the default is min(logical CPUs in the affinity mask, cgroup CPU
// limit). That is two inputs, not three: the runtime counts logical CPUs with
// sched_getaffinity, so the CPU count and the affinity mask are one number.
// Only the cgroup limit is rounded UP and floored at 2; an affinity mask of
// one CPU (taskset -c 0) still gives 1. See runtime/cgroup_linux.go:86-94.
// It is also re-checked periodically, by a real goroutine you can see in
// `GODEBUG=scheddetail=1` output, named "GOMAXPROCS updater (idle)".
//
// Build first, then run the binary. `go run` under taskset or GODEBUG also
// applies them to the go command and the compiler.
//
//	env -u GOROOT go build -o /tmp/bin/06-gomaxprocs ./06-gomaxprocs
//	/tmp/bin/06-gomaxprocs
//	taskset -c 0 /tmp/bin/06-gomaxprocs          # affinity mask of 1
//	GOMAXPROCS=3 /tmp/bin/06-gomaxprocs          # explicit: disables updates
//	GODEBUG=containermaxprocs=0 /tmp/bin/06-gomaxprocs
//
// Under GOMAXPROCS=3 the last line does not print 3. SetDefaultGOMAXPROCS
// ignores the GOMAXPROCS environment variable, so it returns to the
// CPU-derived default (20 on the 20-CPU WSL box).
package main

import (
	"fmt"
	"runtime"
	"runtime/metrics"
)

func metric(name string) uint64 {
	s := []metrics.Sample{{Name: name}}
	metrics.Read(s)
	return s[0].Value.Uint64()
}

func main() {
	fmt.Printf("runtime.NumCPU()          = %d\n", runtime.NumCPU())
	fmt.Printf("runtime.GOMAXPROCS(0)     = %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("/sched/gomaxprocs:threads = %d\n", metric("/sched/gomaxprocs:threads"))
	fmt.Printf("/sched/threads/total      = %d\n", metric("/sched/threads/total:threads"))

	// Setting GOMAXPROCS explicitly disables the automatic updates.
	// SetDefaultGOMAXPROCS (Go 1.25) re-arms them.
	old := runtime.GOMAXPROCS(1)
	fmt.Printf("\nafter GOMAXPROCS(1): now %d (was %d): automatic updates now OFF\n",
		runtime.GOMAXPROCS(0), old)
	runtime.SetDefaultGOMAXPROCS()
	fmt.Printf("after SetDefaultGOMAXPROCS(): %d, updates re-armed\n", runtime.GOMAXPROCS(0))
}
