// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package probe

import (
	"reflect"

	"course/w02/practice/feed"
)

// ReflectSelect merges any number of inputs with one reflect.Select loop: the
// generated loops' shape, with the cases built at run time. A case whose Chan
// is the zero reflect.Value is ignored, the reflect counterpart of a nil
// channel, so a closed input is switched off by clearing its Chan.
func ReflectSelect(ins ...<-chan feed.Item) <-chan feed.Item {
	cases := make([]reflect.SelectCase, len(ins))
	for k, in := range ins {
		cases[k] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(in)}
	}
	out := make(chan feed.Item)
	go func() {
		defer close(out)
		for open := len(ins); open > 0; {
			k, v, ok := reflect.Select(cases)
			if !ok {
				cases[k].Chan = reflect.Value{}
				open--
				continue
			}
			out <- v.Interface().(feed.Item)
		}
	}()
	return out
}
