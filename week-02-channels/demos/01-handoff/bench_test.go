package handoff

import (
	"sync"
	"testing"
)

// Ping-pong across an unbuffered channel: every send is a direct handoff.
func BenchmarkUnbuffered(b *testing.B) { pingpong(b, 0) }

// Same protocol, one slot of buffer: the send returns without a rendezvous.
func BenchmarkBuffered1(b *testing.B) { pingpong(b, 1) }

func pingpong(b *testing.B, size int) {
	ping := make(chan int, size)
	pong := make(chan int, size)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ping {
			pong <- v
		}
		close(pong)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ping <- i
		<-pong
	}
	b.StopTimer()
	close(ping)
	wg.Wait()
}

// One producer, one consumer, streaming through a buffer of varying size.
func BenchmarkStream0(b *testing.B)    { stream(b, 0) }
func BenchmarkStream1(b *testing.B)    { stream(b, 1) }
func BenchmarkStream64(b *testing.B)   { stream(b, 64) }
func BenchmarkStream1024(b *testing.B) { stream(b, 1024) }

func stream(b *testing.B, size int) {
	ch := make(chan int, size)
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	b.StopTimer()
	close(ch)
	<-done
}
