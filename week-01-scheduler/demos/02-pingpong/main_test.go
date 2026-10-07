// Demo 2: runnext, or why two goroutines talking to each other are scheduled
// as a unit.
//
// This is Austin Clements' benchmark from CL 9288 (April 2015), adapted from
// src/runtime/proc_test.go in the Go distribution (BSD-3-Clause, (c) The Go
// Authors). Two goroutines ping-pong on unbuffered channels while a third
// goroutine hogs the CPU and never yields voluntarily. All three share one P.
//
// Before CL 9289 a ready'd goroutine went to the TAIL of the local run queue,
// so every ping-pong round-trip waited behind the hog's entire time slice.
// Austin measured 684287 ns/op. After adding one new field, p.runnext, and the
// scheduler code that fills and drains it, he measured 825 ns/op, a 99.88%
// improvement.
//
// Run:
//
//	go test -run=X -bench=PingPong -benchtime=200000x ./02-pingpong
//
// Expect BenchmarkPingPongHog to be within a small factor of BenchmarkPingPong.
// In 2015, before runnext, PingPongHog took 684287 ns/op; after the fix, 825.
package pingpong

import (
	"runtime"
	"testing"
)

func benchmarkPingPong(b *testing.B, hog bool) {
	if b.N == 0 {
		return
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

	stop, done := make(chan bool), make(chan bool)
	if hog {
		// A CPU hog with no cooperative scheduling point of its own.
		go func() {
			for {
				select {
				case <-stop:
					done <- true
					return
				default:
				}
			}
		}()
	}

	ping, pong := make(chan bool), make(chan bool)
	go func() {
		for range b.N {
			pong <- <-ping
		}
		close(stop)
		done <- true
	}()
	go func() {
		for range b.N {
			ping <- <-pong
		}
		done <- true
	}()

	b.ResetTimer()
	ping <- true // start
	<-stop
	b.StopTimer()
	<-ping // let the last ponger exit
	<-done
	<-done
	if hog {
		<-done
	}
}

func BenchmarkPingPong(b *testing.B)    { benchmarkPingPong(b, false) }
func BenchmarkPingPongHog(b *testing.B) { benchmarkPingPong(b, true) }
