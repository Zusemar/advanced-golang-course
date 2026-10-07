package main

import (
	"syscall"
	"time"
)

// blockingSyscall parks the goroutine inside nanosleep(2). The runtime cannot
// hand this off to the netpoller, so the M stays in the kernel for the whole
// duration and sysmon must retake the P.
func blockingSyscall(d time.Duration) {
	ts := syscall.NsecToTimespec(int64(d))
	for {
		err := syscall.Nanosleep(&ts, &ts)
		if err != syscall.EINTR {
			return
		}
		// EINTR: async preemption signals interrupt sleeps (Go 1.14+).
	}
}
