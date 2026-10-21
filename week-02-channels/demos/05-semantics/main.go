package main

import (
	"fmt"
	"time"
)

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%-42s panic: %v\n", name, r)
			return
		}
		fmt.Printf("%-42s ok\n", name)
	}()
	f()
}

func main() {
	try("recv from closed chan", func() {
		ch := make(chan int)
		close(ch)
		v, ok := <-ch
		fmt.Printf("%-42s v=%d ok=%v\n", "  ->", v, ok)
	})
	try("recv drains buffer, then reports closed", func() {
		ch := make(chan int, 2)
		ch <- 7
		close(ch)
		v1, ok1 := <-ch
		v2, ok2 := <-ch
		fmt.Printf("%-42s v=%d ok=%v then v=%d ok=%v\n", "  ->", v1, ok1, v2, ok2)
	})
	try("send on closed chan", func() {
		ch := make(chan int, 1)
		close(ch)
		ch <- 1
	})
	try("close of closed chan", func() {
		ch := make(chan int)
		close(ch)
		close(ch)
	})
	try("close of nil chan", func() {
		var ch chan int
		close(ch)
	})
	try("nil chan in select with default", func() {
		var ch chan int
		select {
		case <-ch:
			fmt.Println("  -> impossible")
		default:
			fmt.Printf("%-42s nil case is never ready; default taken\n", "  ->")
		}
	})
	try("len/cap of a closed chan", func() {
		ch := make(chan int, 4)
		ch <- 1
		close(ch)
		fmt.Printf("%-42s len=%d cap=%d\n", "  ->", len(ch), cap(ch))
	})
	try("len/cap of a time.Timer chan", func() {
		t := time.NewTimer(time.Millisecond)
		time.Sleep(20 * time.Millisecond)
		fmt.Printf("%-42s len=%d cap=%d (buffer holds a value)\n", "  ->", len(t.C), cap(t.C))
	})
}
