// SPDX-License-Identifier: Apache-2.0
//
// The public battery. Do not edit: it is replaced by the staff battery at
// grading, which checks the same contract more thoroughly.
//
//	go test -race -count=20 -cpu=1,2,8 .
//	go test .
//
// The first skips TestFloor, which lives in public_norace_test.go and is built
// only without -race; the second runs everything, TestFloor included.

package parallelmap_test

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	parallelmap "deepgo/week-01-parallelmap"
)

func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

func TestOrder(t *testing.T) {
	out, err := parallelmap.Map(context.Background(), seq(100), 8, func(_ context.Context, i int) (int, error) {
		time.Sleep(time.Duration(7-i%8) * 50 * time.Microsecond)
		return i * i, nil
	})
	if err != nil || len(out) != 100 {
		t.Fatalf("got len %d, err %v; want 100 results, nil", len(out), err)
	}
	for i, v := range out {
		if v != i*i {
			t.Fatalf("out[%d] = %d, want %d", i, v, i*i)
		}
	}
}

func TestBoundaries(t *testing.T) {
	f := func(_ context.Context, i int) (int, error) { return i, nil }
	if out, err := parallelmap.Map(context.Background(), nil, 4, f); err != nil || len(out) != 0 {
		t.Errorf("empty input: got %v, %v; want empty, nil", out, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, n := range []int{10, 0} {
		if out, err := parallelmap.Map(ctx, seq(n), 4, f); out != nil || err != context.Canceled {
			t.Errorf("cancelled ctx, len %d: got %v, %v; want nil, context.Canceled", n, out, err)
		}
	}
	for _, w := range []int{0, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("workers=%d did not panic", w)
				}
			}()
			parallelmap.Map(context.Background(), seq(10), w, f)
		}()
	}
}

func TestConcurrency(t *testing.T) {
	const w = 4
	var cur, peak atomic.Int32
	parallelmap.Map(context.Background(), seq(100), w, func(_ context.Context, i int) (int, error) {
		c := cur.Add(1)
		for p := peak.Load(); c > p && !peak.CompareAndSwap(p, c); p = peak.Load() {
		}
		time.Sleep(200 * time.Microsecond)
		cur.Add(-1)
		return i, nil
	})
	if p := peak.Load(); p > w {
		t.Fatalf("%d calls ran at once, want at most %d", p, w)
	}
}

func TestErrorCancels(t *testing.T) {
	boom := errors.New("boom")
	var lag time.Duration
	for range 3 {
		var calls atomic.Int32
		var failedAt time.Time
		done := make(chan struct{})
		var out []int
		var err error
		go func() {
			defer close(done)
			out, err = parallelmap.Map(context.Background(), seq(1000), 4, func(ctx context.Context, i int) (int, error) {
				if calls.Add(1) == 1 {
					time.Sleep(5 * time.Millisecond)
					failedAt = time.Now()
					return 0, boom
				}
				<-ctx.Done()
				return 0, ctx.Err()
			})
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Map did not return after an error: is the context passed to f cancelled?")
		}
		if out != nil || !errors.Is(err, boom) {
			t.Fatalf("got %v, %v; want nil, %v", out, err, boom)
		}
		if c := calls.Load(); c > 500 {
			t.Fatalf("f was called %d times after the first error; Map must stop starting calls", c)
		}
		if lag = time.Since(failedAt); lag < 100*time.Millisecond {
			return
		}
	}
	t.Fatalf("Map returned %v after the first error; the grader allows 100ms", lag)
}

func TestFirstErrorInTime(t *testing.T) {
	boom, late := errors.New("boom"), errors.New("late")
	_, err := parallelmap.Map(context.Background(), seq(4), 4, func(ctx context.Context, i int) (int, error) {
		if i == 2 {
			return 0, boom
		}
		select {
		case <-ctx.Done():
			return 0, late
		case <-time.After(50 * time.Millisecond):
			return i, nil
		}
	})
	if !errors.Is(err, boom) {
		t.Fatalf("got %v, want %v: return the error that happened first, not the lowest index", err, boom)
	}
}

func TestParentCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	out, err := parallelmap.Map(ctx, seq(1000), 4, func(_ context.Context, i int) (int, error) {
		if calls.Add(1) == 10 {
			cancel()
		}
		time.Sleep(50 * time.Microsecond)
		return i, nil
	})
	if out != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("got len %d, %v; want nil, context.Canceled even though f ignores its ctx", len(out), err)
	}
}

func TestNoGoroutinesLeft(t *testing.T) {
	before := runtime.NumGoroutine()
	parallelmap.Map(context.Background(), seq(100), 8, func(_ context.Context, i int) (int, error) {
		if i == 50 {
			return 0, errors.New("boom")
		}
		return i, nil
	})
	for deadline := time.Now().Add(time.Second); runtime.NumGoroutine() > before; time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("%d goroutines still running after Map returned", runtime.NumGoroutine()-before)
		}
	}
}
