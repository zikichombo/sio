// +build cgo
// +build ignore
//
// we need gcc 4.9 on travis for this, which I don't have time to set up just yet.

package libsio

// #include "rb.h"
import "C"
import (
	"sync/atomic"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// Type Rb is a ringbuffer linking Go and C for sound i/o
// to callback based C APIs, tuned to the case where the the
// callback occurs on a different thread.
//
// A user of Rb in implementing a host.Entry should map
// the C callback API to the callback functions in rb.h
// for some Rb * object.  She should then create a Go Rb
// with the C Rb * object and then the result will implement
// sound.{Source,Sink,Duplex}.
//
// Other work will be necessary such as calling the C API
// specific functions for setting up and closing the callback.
// This is not in the scope of rb.
type Rb struct {
	sound.Form
	c   *C.Rb
	sco sample.Codec
	ds  [][]float64
	bsz int
	// TBD: leftover if d isn't multiple of b
}

func (r *Rb) Close() error {
	C.freeRb(r.c)
	return nil
}

func (r *Rb) Receive(d []float64) (int, error) {
	N := len(d)
	b := int(r.c.bufSz)
	start := 0
	var end int
	var sl []float64
	addr := (*uint32)(&r.c.size) // why is this cast necessary?  Dangerous?
	var sz uint32
	var cbBuf []byte
	for start < N {
		// get the slice at buffer size
		end = start + b
		if end >= N {
			end = N - 1
		}
		sl = d[start:end]
		sz = 0
		for sz == 0 {
			sz = atomic.LoadUint32(addr)
		}
		// "render"
		cbBuf = (*[1 << 30]byte)(unsafe.Pointer(C.getIn(r.c)))[:r.c.bufSz]
		r.sco.Decode(sl, cbBuf)

		// set state
		C.incRi(r.c)
		// atomic decr
		for {
			sz = atomic.LoadUint32(addr)
			if atomic.CompareAndSwapUint32(addr, sz, sz-1) {
				break
			}
		}
	}
	return 0, nil
}

func (r *Rb) Send(d []float64) error {
	return nil
}

func (r *Rb) SendReceive(out, in []float64) (int, error) {
	return 0, nil
}
