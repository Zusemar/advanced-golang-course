package main

import (
	"fmt"
	"time"
)

type pair struct{ a, b int }

var p pair

func part1() {
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			for i := range 1000 {
				if i&1 == 0 {
					p = pair{1, 1}
				} else {
					p = pair{2, 2}
				}
			}
		}
	}()

	torn, reads := 0, 0
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		for range 1000 {
			v := p
			reads++
			if v.a != v.b {
				torn++
			}
		}
	}
	close(stop)
	fmt.Printf("part 1  struct{a,b int}: %d reads, %d observed a != b\n", reads, torn)
	fmt.Println("        Only {1,1} and {2,2} were ever written.")
}

// The two values a string header ever holds here: 32 bytes, and 1 byte. A torn
// read pairs the length of one with the data pointer of the other.
const (
	long  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 32 bytes
	short = "b"                                // 1 byte
)

var s string

func part2() {
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			for i := range 1000 {
				if i&1 == 0 {
					s = long
				} else {
					s = short
				}
			}
		}
	}()

	torn, reads := 0, 0
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		for range 1000 {
			v := s
			reads++
			switch len(v) {
			case len(long):
				// Length from long, and if the pointer came from short this
				// reads the byte at offset 31, 30 bytes past the end of a
				// one-byte string. In bounds as far as the program is
				// concerned; out of bounds in fact.
				if v[len(long)-1] != long[len(long)-1] {
					torn++
				}
			case len(short):
				if v[0] != short[0] {
					torn++
				}
			}
		}
	}
	close(stop)
	fmt.Printf("part 2  string header:   %d reads, %d had a length from one\n", reads, torn)
	fmt.Println("        string and a data pointer from the other.")
}

func main() {
	part1()
	part2()
}
