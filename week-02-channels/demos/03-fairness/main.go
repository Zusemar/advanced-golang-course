package main

import "fmt"

func main() {
	const N = 300000
	a := make(chan int, 1)
	b := make(chan int, 1)
	c := make(chan int, 1)
	counts := make([]int, 3)

	for i := 0; i < N; i++ {
		// All three cases are ready on every iteration.
		a <- 1
		b <- 1
		c <- 1
		select {
		case <-a:
			counts[0]++
		case <-b:
			counts[1]++
		case <-c:
			counts[2]++
		}
		// drain the two that lost
		for _, ch := range []chan int{a, b, c} {
			select {
			case <-ch:
			default:
			}
		}
	}
	fmt.Printf("three always-ready cases, %d iterations\n", N)
	for i, n := range counts {
		fmt.Printf("  case %d: %7d  (%.2f%%)\n", i, n, 100*float64(n)/float64(N))
	}
}
