// SPDX-License-Identifier: Apache-2.0

package shardedcounter

// Tier 2 only: optional, pass/fail. Leave this file as it is if you do not
// attempt Tier 2. See README.md, "Tier 2".

// Stack is a lock-free LIFO stack (a Treiber stack), safe for concurrent use
// by multiple goroutines. Use NewStack to make one. It takes no lock: this
// file imports nothing from sync except sync/atomic.
type Stack[T any] struct {
	// Your fields.
}

// NewStack returns an empty stack.
func NewStack[T any]() *Stack[T] {
	panic("shardedcounter: NewStack is not implemented")
}

// Push puts v on top of the stack.
func (s *Stack[T]) Push(v T) {
	panic("shardedcounter: Push is not implemented")
}

// Pop removes the value on top of the stack and returns it with ok true. On an
// empty stack it returns the zero T and ok false, and does not block.
func (s *Stack[T]) Pop() (v T, ok bool) {
	panic("shardedcounter: Pop is not implemented")
}
