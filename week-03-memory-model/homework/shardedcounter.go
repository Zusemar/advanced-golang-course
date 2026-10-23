// SPDX-License-Identifier: Apache-2.0

// Package shardedcounter is a counter that many goroutines add to at once.
//
// The doc comments in this file are the contract. Where they and README.md
// disagree, this file wins; the grader is written against it.
package shardedcounter

// ShardedCounter is an int64 counter that is safe for concurrent use by
// multiple goroutines. Use New to make one; the zero value is not required to
// work. A ShardedCounter must not be copied after first use.
type ShardedCounter struct {
	// Your fields.
}

// New returns a counter whose Sum is 0.
func New() *ShardedCounter {
	panic("shardedcounter: New is not implemented")
}

// Add adds delta, which may be negative, to the counter. It is safe to call
// from any number of goroutines at once, concurrently with Sum. Overflow wraps,
// as int64 addition does.
func (c *ShardedCounter) Add(delta int64) {
	panic("shardedcounter: Add is not implemented")
}

// Sum returns the counter's value. When no Add runs concurrently with it, that
// is exactly the total of every delta added so far.
//
// TODO: state what Sum guarantees while Adds run concurrently with it. This
// comment is graded: see README.md, "The snapshot statement".
func (c *ShardedCounter) Sum() int64 {
	panic("shardedcounter: Sum is not implemented")
}
