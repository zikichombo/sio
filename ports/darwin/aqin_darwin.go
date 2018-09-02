// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package darwin

import (
	"fmt"
	"sync"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// #cgo LDFLAGS: -framework AudioToolbox -framework CoreFoundation
//
// #include <AudioToolbox/AudioQueue.h>
// #include <CoreAudio/CoreAudioTypes.h>
// #include <CoreFoundation/CFRunLoop.h>
// void inputCallback(
//		void *goInput,
//		AudioQueueRef q,
//		AudioQueueBufferRef buf,
//		const AudioTimeStamp* unusedTs,
//		UInt32 numPackets,
//		const AudioStreamPacketDescription* unusedPD);
import "C"

const (
	maxAqins = 256
)

type aqin struct {
	mu   sync.Mutex
	once sync.Once
	id   int
	ch   chan *Packet
	cch  chan struct{}
	aq
	running bool
	n       int
}

func newAqin(v sound.Form, co sample.Codec, nFrames int) (*aqin, error) {
	id := <-_inaqNew
	q := &_inaqs[id]
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

//export inGoCallback
func inGoCallback(id int, qb *C.struct_AudioQueueBuffer, d *C.char, sz int) {
	q := &_inaqs[id]
	q.callback(qb, d, sz)
}

func (q *aqin) callback(qb *C.struct_AudioQueueBuffer, d *C.char, sz int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.ch == nil {
		return
	}
	pkt := &q.gBufs[q.gbp]
	q.gbp++
	if q.gbp == len(q.gBufs) {
		q.gbp = 0
	}
	if !q.running { // can happen if multiple callbacks are queued
		return
	}
	dsz := q.codec.Bytes()
	slice := (*[1 << 30]byte)(unsafe.Pointer(d))[:sz]
	N := sz / dsz
	pkt.N = q.n
	pkt.D = pkt.D[:N]
	q.n += N / q.Channels()
	q.codec.Decode(pkt.D, slice)
	select {
	case q.ch <- pkt:
	case <-q.cch:
		// else stopped, enqueue the buffer so it will be there if start
		// is subsequently called
		q.running = false
		return
	}
	st, _ := C.AudioQueueEnqueueBuffer(q.qRef, qb, 0, nil)
	if e := caStatus(st); e != nil {
		fmt.Printf("error enqueuing in callback: %s\n", e)
	}
}

func (q *aqin) start() error {
	err := q.aq.start()
	if err != nil {
		return err
	}
	q.running = true
	return nil
}

func (q *aqin) Close() error {
	q.cch <- struct{}{}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.once.Do(func() { close(q.ch) })
	q.running = false
	if err := q.stop(); err != nil {
		fmt.Printf("error stopping inaq: %s\n", err)
	}
	q.aqClose()
	_inaqFree <- q.id
	return nil
}

func (q *aqin) C() <-chan *Packet {
	return q.ch
}

func (q *aqin) init(v sound.Form, co sample.Codec, bufSize int) error {
	q.initFormat(v, co)
	q.initBufs(bufSize)
	var err error
	defer func() {
		if err == nil {
			return
		}
		q.Close()
	}()
	st, _ := C.AudioQueueNewInput(
		&q.fmt,
		(*[0]byte)(unsafe.Pointer(C.inputCallback)),
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
	q.enqueueBufs()
	q.ch = make(chan *Packet, 1)
	q.cch = make(chan struct{}, 1)
	q.running = false
	q.n = 0
	return nil
}

// globals so we can not have go pointers to go pointers in c.
// instead we refer to ids.
var _inaqs [maxAqins]aqin
var _inaqFree chan int
var _inaqNew chan int

// set up process of ids and free list.
// now, since each id refers to fixed place in memory
// we can use it without locking in the sound data
// processing loop.
func init() {
	_inaqFree = make(chan int, maxAqins)
	_inaqNew = make(chan int)
	for i := 0; i < maxAqins; i++ {
		_inaqs[i].id = i
		_inaqFree <- i
	}
	go func() {
		var f int
		for {
			f = <-_inaqFree
			_inaqNew <- f
		}
	}()
}
