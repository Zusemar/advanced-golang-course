// SPDX-License-Identifier: Apache-2.0
//
// The public battery. Do not edit: it is replaced by the staff battery at
// grading, which checks the same contract more thoroughly.
//
//	go test -race -count=20 -cpu=1,2,8 .
//	go test .
//
// The first skips TestSlowSubscriber, which lives in public_norace_test.go and
// is built only without -race; the second runs everything, it included.

package broadcaster_test

import (
	"testing"
	"time"

	broadcaster "deepgo/week-02-broadcaster"
)

// pubReader collects what sub receives; each value is also sent on acks.
func pubReader(sub *broadcaster.Subscription[int]) (acks chan int, got func() []int) {
	acks = make(chan int, 1024)
	done := make(chan struct{})
	var vals []int
	go func() {
		defer close(done)
		for v := range sub.C() {
			vals = append(vals, v)
			acks <- v
		}
	}()
	return acks, func() []int {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			panic("a subscription's channel was not closed")
		}
		return vals
	}
}

func await(t *testing.T, acks chan int, want int) {
	t.Helper()
	select {
	case v := <-acks:
		if v != want {
			t.Fatalf("received %d, want %d", v, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("message %d was never delivered", want)
	}
}

func TestDelivery(t *testing.T) {
	b := broadcaster.New[int]()
	var acks []chan int
	var gots []func() []int
	for range 3 {
		a, g := pubReader(b.Subscribe(2))
		acks, gots = append(acks, a), append(gots, g)
	}
	for i := range 50 {
		b.Publish(i)
		for _, a := range acks {
			await(t, a, i)
		}
	}
	b.Close()
	for _, g := range gots {
		if n := len(g()); n != 50 {
			t.Fatalf("got %d messages, want 50", n)
		}
	}
}

func TestSubscribeLater(t *testing.T) {
	b := broadcaster.New[int]()
	early, _ := pubReader(b.Subscribe(1))
	for i := range 10 {
		b.Publish(i)
		await(t, early, i)
	}
	late, got := pubReader(b.Subscribe(1))
	for i := 10; i < 20; i++ {
		b.Publish(i)
		await(t, early, i)
		await(t, late, i)
	}
	b.Close()
	if g := got(); len(g) != 10 || g[0] != 10 {
		t.Fatalf("a late subscriber got %v, want 10..19", g)
	}
}

func TestUnsubscribe(t *testing.T) {
	b := broadcaster.New[int]()
	sub := b.Subscribe(4)
	b.Publish(1)
	sub.Unsubscribe()
	for {
		select {
		case _, ok := <-sub.C():
			if ok {
				continue
			}
		default:
			t.Fatal("C was not closed when Unsubscribe returned")
		}
		break
	}
	b.Publish(2)
	sub.Unsubscribe()
	b.Close()
}

func TestSubscribeBufferBelowOne(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Subscribe(0) did not panic")
		}
	}()
	broadcaster.New[int]().Subscribe(0)
}

func TestClose(t *testing.T) {
	b := broadcaster.New[int]()
	subs := []*broadcaster.Subscription[int]{b.Subscribe(1), b.Subscribe(3)}
	b.Publish(1)
	b.Close()
	for _, s := range append(subs, b.Subscribe(1)) {
		closed := false
		for !closed {
			select {
			case _, ok := <-s.C():
				closed = !ok
			default:
				t.Fatal("a channel was not closed when Close returned")
			}
		}
	}
}
