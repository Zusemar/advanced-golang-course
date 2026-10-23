package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

const (
	workers = 4
	ops     = 20_000
	nodes   = 8
)

type node struct {
	val  int
	next *node
}

type stack struct {
	top atomic.Pointer[node]
}

func (s *stack) push(n *node) {
	for {
		old := s.top.Load()
		n.next = old
		if s.top.CompareAndSwap(old, n) {
			return
		}
	}
}

func (s *stack) pop(widen bool) *node {
	for {
		old := s.top.Load()
		if old == nil {
			return nil
		}
		next := old.next
		if widen {
			runtime.Gosched() // widen the ABA window (see the note below)
		}
		if s.top.CompareAndSwap(old, next) {
			return old
		}
	}
}

// freelist hands the same node addresses back out. It is deliberately
// mutex-protected: the structure under test is the stack, not this.
type freelist struct {
	mu sync.Mutex
	ns []*node
}

func (f *freelist) put(n *node) {
	f.mu.Lock()
	f.ns = append(f.ns, n)
	f.mu.Unlock()
}

func (f *freelist) take() *node {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ns) == 0 {
		return nil
	}
	n := f.ns[len(f.ns)-1]
	f.ns = f.ns[:len(f.ns)-1]
	return n
}

// walk counts the nodes reachable from the top of the stack and reports
// whether the chain closed on itself.
func walk(s *stack) (count int, cycle bool) {
	seen := make(map[*node]bool, nodes)
	for p := s.top.Load(); p != nil; p = p.next {
		if seen[p] {
			return len(seen), true
		}
		seen[p] = true
	}
	return len(seen), false
}

// run starts four workers doing 20,000 pop/push pairs each over an 8-node
// stack. With recycle=true a popped node's *address* is put on a free list and
// pushed back; with recycle=false every push allocates a fresh node and the
// garbage collector owns the old one.
func run(recycle, widen bool) string {
	var s stack
	var f freelist
	for i := range nodes {
		s.push(&node{val: i})
	}

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for range ops {
				n := s.pop(widen)
				if n == nil {
					continue // the stack lost its nodes; the walk below will say so
				}
				if recycle {
					f.put(n)
					m := f.take()
					if m == nil {
						m = &node{val: n.val}
					}
					s.push(m)
				} else {
					s.push(&node{val: n.val})
				}
			}
		})
	}
	wg.Wait()

	count, cycle := walk(&s)
	if count == nodes && !cycle {
		return "stack intact"
	}
	return fmt.Sprintf("CORRUPT: %d nodes lost, cycle=%v", nodes-count, cycle)
}

func main() {
	fmt.Printf("%d workers, %d ops each, %d nodes on the stack throughout\n\n",
		workers, ops, nodes)
	for _, c := range []struct {
		name    string
		recycle bool
		widen   bool
	}{
		{"fresh nodes (GC owns them)", false, false},
		{"fresh nodes, widened window", false, true},
		{"recycled nodes", true, false},
		{"recycled nodes, widened window", true, true},
	} {
		fmt.Printf("%-32s %s\n", c.name, run(c.recycle, c.widen))
	}
}
