package falsesharing

import "unsafe"

// ptrAt returns a pointer to byte offset off in buf, after aligning the base
// to 64 bytes so that stride arithmetic is measured against real cache-line
// boundaries rather than an arbitrary allocation offset.
func ptrAt(buf []byte, off int) unsafe.Pointer {
	base := uintptr(unsafe.Pointer(&buf[0]))
	pad := (64 - base%64) % 64
	return unsafe.Pointer(&buf[int(pad)+off])
}
