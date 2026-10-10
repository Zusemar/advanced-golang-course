// SPDX-License-Identifier: Apache-2.0

// Package parallelmap applies a function to every element of a slice, in
// parallel, with a bound on how many calls run at once.
package parallelmap

import "context"

// Map calls f on every element of in and returns the results in order:
// out[i] is the result of f(ctx, in[i]). f is called exactly once per element.
//
// At most workers calls of f run at once, and Map keeps workers calls running
// as long as there are elements it has not started. The number of goroutines
// Map starts depends on workers, not on len(in).
//
// Every call of f gets a context derived from ctx. If a call returns an
// error, Map cancels that context, starts no more calls, and returns nil and
// the error. If several calls fail, Map returns the error that happened first
// in time. If ctx is cancelled, Map stops the same way and returns nil and
// ctx.Err(), or the first error from f if there was one. If ctx is cancelled
// after every call has already succeeded, Map may return either the results
// or ctx.Err().
//
// Map returns promptly once it stops, provided f returns when its context is
// cancelled. When Map returns, no call of f is still running and no goroutine
// it started is still alive.
//
// Edge cases, checked in this order:
//   - workers < 1: Map panics without calling f.
//   - ctx is already done: Map returns nil and ctx.Err() without calling f.
//   - len(in) == 0: Map returns an empty slice (nil is fine) and nil.
//
// workers > len(in) is allowed. ctx must not be nil, and f must not panic.
func Map[T, R any](ctx context.Context, in []T, workers int, f func(context.Context, T) (R, error)) ([]R, error) {
	panic("parallelmap: Map is not implemented")
}
