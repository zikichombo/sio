// +build cgo

package libsio

// #cgo CFLAGS: -std=c11
// #include "cb.h"
import "C"
import (
	"io"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/cil"
	"zikichombo.org/sound/sample"
)

// Type Cb is a type linking Go and C for sound i/o to callback based C APIs,
// tuned to the case where the callback occurs on a different thread.
//
// Cb must be used in conjunction with corresponding libsio/cb.h
// C API to be useful.  There are other system requirements and considerations
// as well:
//
//  1. The synchronisation mechanism here assumes that at most one C callback
//  thread is executing a callback at a time.
//
//  2. The C API must accept configuration by buffer size and
//  never present the user with a buffer which exceeds this size.
//
//  3. If the C API is unable to regularly provide the requested buffer
//  size in 2, then
//
//    - for input, there will be latency and CPU overhead
//    - for output, the C API must allow the callback to inform the
//      underlying system of the actual number of frames provided, even
//      for non EOF conditions.  Normally, this means the C API has
//      latency associated with alignment.
//
//  4. For best reliability, the Go code should be run on a thread with the same
//     priority as the C API.
//
//  5. A Multicore system will be more reliable and give better performance than
//     a single core system because the Go code will be on a different thread
//     than the C API callback, and there are no syscalls to tell the OS to
//     switch threads.
//
// If the above requirements and considerations are addressed, then to
// implement host.Entry using Cb, one should map the C callback API to the
// callback functions in cb.h for some Cb * object.  One should then create a
// Go Cb with the C Cb * object and then the result will implement
// sound.{Source,Sink,Duplex}, according to how the callback is mapped. The
// test case in libsio.cb_test.go shows an example for a dummy C API.
//
// Other work will be necessary such as calling the C API specific functions
// for setting up and closing the callback.  This is not in the scope of cb.
type Cb struct {
	sound.Form
	inForm, outForm sound.Form // only for duplex

	c   *C.Cb
	sco sample.Codec
	bsz int
	// zc uses channel deinterleaved for processing, most hardware uses interleaved.
	// il provides adapter functionality.
	il *cil.T

	// just in case the underlying cb api gets out of sync w.r.t. buffer sizes
	// we need to be able to handle it gracefully.  If it happens, it can increase
	// latency and cpu overhead, there is nothing that can be done as any regular alignment
	// of bursts of irregular length data will have this effect.
	over []float64

	// time tracking
	lastTime time.Time
	bufDur   time.Duration
}

func NewCb(v sound.Form, sco sample.Codec, b int) *Cb {
	return &Cb{
		Form:   v,
		sco:    sco,
		bsz:    b,
		il:     cil.New(v.Channels(), b),
		over:   make([]float64, 0, b),
		c:      C.newCb(C.int(b)),
		bufDur: time.Duration(b) * v.SampleRate().Period()}
}

const (
	// amount of slack we give between ask for wake up and
	// pseudo-spin.  guestimated for general OS scheduling
	// latency of worst case 1 preempting task + general Go GC
	// latency.
	sleepSlack = 5 * time.Millisecond
)

func (r *Cb) Close() error {
	C.closeCb(r.c)
	C.freeCb(r.c)
	return nil
}

// Receive is as in sound.Source.Receive
func (r *Cb) Receive(d []float64) (int, error) {
	N := len(d)
	nC := r.Channels()
	if N%nC != 0 {
		return 0, sound.ErrChannelAlignment
	}
	nF := N / nC
	b := r.bsz
	if nF%b != 0 {
		return 0, sound.ErrFrameAlignment
	}
	start := 0
	if len(r.over) != 0 {
		copy(d, r.over)
		start = len(r.over) / nC
		r.over = r.over[:0]
	}
	r.maybeSleep()

	var sl []float64 // per cb subslice of d
	addr := (*uint32)(unsafe.Pointer(&r.c.inGo))
	bps := r.sco.Bytes()
	var nf, onf int  // frame counter and overlap frame count
	var cbBuf []byte // cast from C pointer callback data
	orgTime := r.lastTime
	for start < nF {
		r.fromC(addr)
		if orgTime == r.lastTime {
			r.lastTime = time.Now()
		}

		nf = int(r.c.inF)
		if nf == 0 {
			r.toC(addr)
			break
		}

		// in case the C cb doesn't align to the buffer size
		if start+nf > nF {
			onf = (start + nf) - nF
			nf = nF - start
		} else {
			onf = 0
		}

		sl = d[start*nC : (start+nf)*nC]
		cbBuf = (*[1 << 30]byte)(unsafe.Pointer(C.getIn(r.c)))[:(nf+onf)*bps*nC]

		r.sco.Decode(sl, cbBuf[:nf*bps*nC])

		// handle overlap
		if onf != 0 {
			r.over = r.over[:onf*nC]
			r.sco.Decode(r.over, cbBuf[nf*bps*nC:])
		}

		r.toC(addr)
		start += nf
	}

	r.il.Deinter(d[:start*nC])
	return start, nil
}

// Send is as in sound.Sink.Send
func (r *Cb) Send(d []float64) error {
	N := len(d)
	nC := r.Channels()
	if N%nC != 0 {
		return sound.ErrChannelAlignment
	}
	nF := N / nC
	b := r.bsz
	if nF%b != 0 {
		return sound.ErrFrameAlignment
	}
	r.il.Inter(d)
	start := 0
	var sl []float64
	addr := (*uint32)(unsafe.Pointer(&r.c.inGo))
	bps := r.sco.Bytes()
	var nf int
	var cbBuf []byte
	orgTime := r.lastTime
	for start < nF {
		r.fromC(addr)
		if orgTime == r.lastTime {
			r.lastTime = time.Now()
		}
		// get the slice at buffer size
		nf = int(r.c.outF)
		if nf == 0 {
			r.toC(addr)
			return io.EOF
		}
		if start+nf > nF {
			nf = nF - start
		}
		sl = d[start*nC : (start+nf)*nC]
		// "render"
		cbBuf = (*[1 << 30]byte)(unsafe.Pointer(C.getIn(r.c)))[:nf*bps*nC]
		r.sco.Encode(cbBuf, sl)
		// tell the API about any truncation that happened.
		r.c.outF = C.int(nf)
		r.toC(addr)
		start += nf
	}
	return nil
}

func (r *Cb) SendReceive(out, in []float64) (int, error) {
	return 0, nil
}

func (r *Cb) maybeSleep() {
	var t time.Time
	if r.lastTime == t { // don't sleep on first call.
		return
	}
	passed := time.Since(r.lastTime)
	if passed+sleepSlack < r.bufDur {
		time.Sleep(r.bufDur - (passed + sleepSlack))
	}
}

func (r *Cb) fromC(addr *uint32) {
	var sz uint32
	for {
		sz = atomic.LoadUint32(addr)
		if sz != 0 {
			return
		}
		runtime.Gosched()
	}
}

func (r *Cb) toC(addr *uint32) {
	var sz uint32
	for {
		sz = atomic.LoadUint32(addr)
		if atomic.CompareAndSwapUint32(addr, sz, sz-1) {
			return
		}
		runtime.Gosched()
	}
}
