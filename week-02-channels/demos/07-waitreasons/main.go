package main

import "time"

func main() {
	var nilch chan int

	unbuf := make(chan int)
	full := make(chan int, 1)
	full <- 1
	a, b := make(chan int), make(chan int)

	go func() { <-nilch }()    // chan receive (nil chan)
	go func() { nilch <- 1 }() // chan send (nil chan)
	go func() { <-unbuf }()    // chan receive
	go func() { full <- 2 }()  // chan send
	go func() { select {} }()  // select (no cases)
	go func() {                // select
		select {
		case <-a:
		case b <- 1:
		}
	}()

	time.Sleep(50 * time.Millisecond)
	select {} // main blocks; the runtime declares deadlock and dumps every goroutine
}
