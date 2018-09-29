// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build cgo

package libsio

// #cgo CFLAGS: -std=c11
// #include "cb.h"
import "C"
import (
	"errors"
	"io"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/cil"
	"zikichombo.org/sound/sample"
)

const (
	// amount of slack we give between ask for wake up and
	// pseudo-spin.  guestimated for general OS scheduling
	// latency of worst case 1 preempting task + general Go GC
	// latency.
	sleepSlack = 5 * time.Millisecond

	// nb of times to try an atomic before defaulting to
	// runtime.Gosched, as the later might or might not on some systems
	// and some circumstances invoke a syscall.
	atomicTryLen = 10

	// max number of tries before we assume something killed
	// the C thread
	atomicTryLim = 100000000
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
	bsz int // in frames
	// zc uses channel deinterleaved for processing, most hardware uses interleaved.
	// il provides adapter functionality.
	il *cil.T

	// just in case the underlying cb api gets out of sync w.r.t. buffer sizes
	// we need to be able to handle it gracefully.  If it happens, it can increase
	// latency and cpu overhead, there is nothing that can be done as any regular alignment
	// of bursts of irregular length data will have this effect.
	over []float64

	// time tracking
	frames   int64
	minCbf   int
	orgTime  time.Time // time of first sample
	frameDur time.Duration
}

// NewCb creates a new Cb for the specified form (channels + sample rate)
// sample codec and buffer size b in frames.
func NewCb(v sound.Form, sco sample.Codec, b int) *Cb {
	fd := v.SampleRate().Period()
	return &Cb{
		Form:     v,
		sco:      sco,
		bsz:      b,
		il:       cil.New(v.Channels(), b),
		over:     make([]float64, 0, b),
		c:        C.newCb(C.int(b)),
		minCbf:   b,
		frameDur: fd}
}

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
		start += len(r.over) / nC
		r.over = r.over[:0]
	}
	var sl []float64 // per cb subslice of d
	addr := (*uint32)(unsafe.Pointer(&r.c.inGo))
	bps := r.sco.Bytes()
	var nf, onf int  // frame counter and overlap frame count
	var cbBuf []byte // cast from C pointer callback data
	for start < nF {
		if err := r.fromC(addr); err != nil {
			return 0, ErrCApiLost
		}

		nf = int(r.c.inF)
		if nf == 0 {
			r.toC(addr)
			if err := r.toC(addr); err != nil {
				return 0, ErrCApiLost
			}
			return 0, io.EOF
		}
		if start == 0 && r.frames == 0 {
			r.setOrgTime(-nf)
		}

		// in case the C cb doesn't align to the buffer size
		if start+nf > nF {
			onf = (start + nf) - nF
			nf = nF - start
		} else {
			onf = 0
		}

		sl = d[start*nC : (start+nf)*nC]
		cbBuf = (*[1 << 30]byte)(unsafe.Pointer(r.c.in))[:(nf+onf)*bps*nC]

		r.sco.Decode(sl, cbBuf[:nf*bps*nC])

		// handle overlap
		if onf != 0 {
			r.over = r.over[:onf*nC]
			r.sco.Decode(r.over, cbBuf[nf*bps*nC:])
		}

		if err := r.toC(addr); err != nil {
			return 0, ErrCApiLost
		}
		start += nf
		r.frames += int64(nf)
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
	for start < nF {
		if err := r.fromC(addr); err != nil {
			return ErrCApiLost
		}
		// get the slice at buffer size
		nf = int(r.c.outF)
		if nf == 0 {
			if err := r.toC(addr); err != nil {
				return ErrCApiLost
			}
			return io.EOF
		}
		if start == 0 && r.frames == 0 {
			r.setOrgTime(0)
		}
		if start+nf > nF {
			nf = nF - start
		}
		sl = d[start*nC : (start+nf)*nC]
		// "render"
		cbBuf = (*[1 << 30]byte)(unsafe.Pointer(r.c.out))[:nf*bps*nC]
		r.sco.Encode(cbBuf, sl)
		// tell the API about any truncation that happened.
		r.c.outF = C.int(nf)
		if err := r.toC(addr); err != nil {
			return ErrCApiLost
		}
		start += nf
		r.frames += int64(nf)
	}
	return nil
}

// TBD
func (r *Cb) SendReceive(out, in []float64) (int, error) {
	return 0, nil
}

// set the time for the first sample.  the time is
// the time that we know the underlying API will have
// access to the first sample (for playback) or
// the latest time that the underlying API could have
// recorded the first sample (for capture).
func (r *Cb) setOrgTime(nf int) {
	d := r.frameDur * time.Duration(nf)
	r.orgTime = time.Now().Add(d)
}

// sleep only if underlying API is regular w.r.t.
// supplied buffer sizes and buffer size is bigger
// than OS latency jitter.
func (r *Cb) maybeSleep() {
	if r.minCbf == 0 || r.frames == 0 {
		return
	}
	trg := r.orgTime.Add(time.Duration(int64(r.minCbf)+r.frames) * r.frameDur)
	deadline := time.Until(trg)
	if deadline <= sleepSlack {
		return
	}
	time.Sleep(deadline - sleepSlack)
}

// ErrCApiLost can be returned if the thread running the C API
// is somehow killed or the hardware causes the callbacks to
// block.
var ErrCApiLost = errors.New("too many atomic tries, C callbacks aren't happening.")

func (r *Cb) fromC(addr *uint32) error {
	r.maybeSleep()

	var sz uint32
	i := 0
	for {
		sz = atomic.LoadUint32(addr)
		if sz != 0 {
			return nil
		}
		i++
		if i%atomicTryLen == 0 {
			if i >= atomicTryLim {
				return ErrCApiLost
			}
			// runtime.Gosched may or may not invoke a syscall if many g's on m
			// use sparingly
			runtime.Gosched()
		}
	}
}

func (r *Cb) toC(addr *uint32) error {
	var sz uint32
	i := 0
	for {
		sz = atomic.LoadUint32(addr)
		if atomic.CompareAndSwapUint32(addr, sz, sz-1) {
			return nil
		}
		i++
		if i%atomicTryLen == 0 {
			if i >= atomicTryLim {
				return ErrCApiLost
			}
			// runtime.Gosched may or may not invoke a syscall if many g's on m
			// use sparingly
			runtime.Gosched()
		}
	}
}
