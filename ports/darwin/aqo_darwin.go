// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package darwin

import (
	"fmt"
	"log"
	"sync"
	"time"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// #cgo LDFLAGS: -framework AudioToolbox -framework CoreFoundation
//
// #include <AudioToolbox/AudioQueue.h>
// #include <CoreAudio/CoreAudioTypes.h>
// #include <CoreFoundation/CFRunLoop.h>
// void outputCallback(
//		void *goOutput,
//		AudioQueueRef q,
//		AudioQueueBufferRef buf);
import "C"

const (
	maxAqos = 256
)

type aqo struct {
	mu  sync.Mutex
	id  int
	chs [2]chan *Packet
	cch chan struct{}
	aq
	running bool
	n       int
}

func newAqo(v sound.Form, co sample.Codec, nFrames int) (*aqo, error) {
	id := <-_aqoNew
	q := &_aqos[id]
	err := q.init(v, co, nFrames)
	if err != nil {
		return nil, err
	}
	if err := q.start(); err != nil {
		q.Close()
		return nil, err
	}
	return q, nil
}

func (q *aqo) start() error {
	q.running = true
	go func() {
		for i := 0; i < 3; i++ {
			qb := q.qBufs[i]
			C.outputCallback(unsafe.Pointer(uintptr(q.id)), q.qRef, qb)
			if q.chs[0] == nil {
				return
			}
		}
		var s *C.struct_AudioTimeStamp
		st, _ := C.AudioQueueStart(q.qRef, s)
		now := time.Now()
		q.gBufs[0].Start = now
		q.gBufs[1].Start = now
		q.gBufs[2].Start = now
		if err := caStatus(st); err != nil {
			fmt.Printf("error starting q: %s\n", err)
		}
	}()
	return nil
}

func (q *aqo) Close() error {
	q.cch <- struct{}{}
	q.mu.Lock()
	defer q.mu.Unlock()
	q._close()
	return nil
}

func (q *aqo) _close() {
	if !q.running {
		return
	}
	q.running = false
	q.stop()
	// TBD: flush
	q.aqClose()
	q.free()
}

func (q *aqo) FillC() <-chan *Packet {
	return q.chs[0]
}

func (q *aqo) PlayC() chan<- *Packet {
	return q.chs[1]
}

func (q *aqo) init(v sound.Form, c sample.Codec, bufSize int) error {
	q.initFormat(v, c)
	q.initBufs(bufSize)
	var err error
	defer func() {
		if err == nil {
			return
		}
		q.Close()
	}()
	st, _ := C.AudioQueueNewOutput(
		&q.fmt,
		(*[0]byte)(unsafe.Pointer(C.outputCallback)),
		unsafe.Pointer(uintptr(q.id)),
		C.CFRunLoopRef(0),
		C.kCFRunLoopCommonModes,
		0,
		&q.qRef)
	err = caStatus(st)
	if err != nil {
		return err
	}
	err = q.allocateBufs(bufSize)
	if err != nil {
		return err
	}
	C.AudioQueueSetParameter(q.qRef, C.kAudioQueueParam_Volume, 1.0)
	q.chs[0] = make(chan *Packet)
	q.chs[1] = make(chan *Packet, 1)
	q.cch = make(chan struct{}, 1)
	q.n = 0
	return nil
}

//export outGoCallback
func outGoCallback(id int, qb *C.struct_AudioQueueBuffer, d *C.char, cap int) {
	q := &_aqos[id]
	q.callback(qb, d, cap)
}

func (q *aqo) callback(qb *C.struct_AudioQueueBuffer, d *C.char, cap int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.chs[0] == nil {
		return
	}
	if !q.running {
		return
	}
	pkt := &q.gBufs[q.gbp]
	pkt.N = q.n
	q.gbp++
	if q.gbp == len(q.gBufs) {
		q.gbp = 0
	}
	p := &pkt.D[0]
	select {
	case q.chs[0] <- pkt:
	case <-q.cch:
		q._close()
		return
	}

	select {
	case in, open := <-q.chs[1]:
		if !open {
			q._close()
		} else {
			if &in.D[0] != p {
				panic("received foreign buffer")
			}
			q.handleBuf(in, qb, d, cap)
		}
	case <-q.cch:
		q._close()
	}
}

func (q *aqo) handleBuf(in *Packet, qb *C.struct_AudioQueueBuffer, d *C.char, cap int) {
	dsz := q.codec.Bytes() //C.sizeof_Float32
	bsz := len(in.D) * dsz
	slice := (*[1 << 30]byte)(unsafe.Pointer(d))[:bsz] //cap]
	q.codec.Encode(slice, in.D)
	qb.mAudioDataByteSize = C.UInt32(bsz)
	// deal with scheduling.
	if in.N < q.n {
		log.Printf("cannot schedule playback non-monotonically: %d < %d\n", in.N, q.n)
		in.N = q.n
	}
	off := in.N - q.n
	var its C.AudioTimeStamp
	its.mSampleTime = C.Float64(float64(in.N))
	its.mFlags = C.kAudioTimeStampSampleTimeValid
	var ots C.AudioTimeStamp
	st, _ := C.AudioQueueEnqueueBufferWithParameters(q.qRef, qb, 0, nil, 0, 0 /* trim end */, 0, nil, &its, &ots)
	if err := caStatus(st); err != nil {
		log.Printf("error enqueueing in output sound queue: %s\n", err)
	}
	if q.n == 0 {
		// tried this but it returns 0 until all buffers have been filled.
		// so you cannot know until too late.
		//log.Printf("host time start %d\n", uint64(ots.mHostTime))
	}
	q.n += len(in.D) / q.Channels()
	q.n += off
	if off > 0 {
		///log.Printf("scheduling %d samples in future\n", off)
	}
}

func (q *aqo) free() {
	close(q.chs[0])
	q.chs[0] = nil
	close(q.chs[1])
	q.chs[1] = nil
	close(q.cch)
	q.cch = nil
	_aqoFree <- q.id
}

// globals so we can not have go pointers to go pointers in c.
// instead we refer to ids.
var _aqos [maxAqos]aqo
var _aqoFree chan int
var _aqoNew chan int

// set up process of ids and free list.
// now, since each id refers to fixed place in memory
// we can use it without locking in the sound data
// processing loop.
func init() {
	_aqoFree = make(chan int, maxAqos)
	_aqoNew = make(chan int)
	for i := 0; i < maxAqos; i++ {
		_aqos[i].id = i
		_aqoFree <- i
	}
	go func() {
		var f int
		for {
			f = <-_aqoFree
			_aqoNew <- f
		}
	}()
}
