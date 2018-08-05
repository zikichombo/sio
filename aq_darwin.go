// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/irifrance/snd"
	"github.com/irifrance/snd/sample"
)

// #cgo LDFLAGS: -framework AudioToolbox -framework CoreFoundation
//
// #include <AudioToolbox/AudioQueue.h>
// #include <CoreAudio/CoreAudioTypes.h>
// #include <CoreFoundation/CFRunLoop.h>
import "C"

type aq struct {
	snd.Form
	codec sample.Codec
	qRef  C.AudioQueueRef
	qBufs [3]C.AudioQueueBufferRef
	gBufs [3]Packet
	gbp   int
	fmt   C.AudioStreamBasicDescription
}

func (q *aq) start() error {
	var s *C.struct_AudioTimeStamp
	st, _ := C.AudioQueueStart(q.qRef, s)
	now := time.Now()
	q.gBufs[0].Start = now
	q.gBufs[1].Start = now
	q.gBufs[2].Start = now
	err := caStatus(st)
	return err
}

func (q *aq) stop() error {
	var f C.Boolean = 0
	st, _ := C.AudioQueueStop(q.qRef, f)
	err := caStatus(st)
	return err
}

func (q *aq) aqClose() {
	var f C.Boolean = 0
	for i := range q.qBufs {
		C.AudioQueueFreeBuffer(q.qRef, q.qBufs[i])
	}
	C.AudioQueueDispose(q.qRef, f)
	q.gBufs[0].D = nil
	q.gBufs[1].D = nil
	q.gBufs[2].D = nil
	q.gbp = 0
}

func (q *aq) allocateBufs(nFrames int) error {
	sz := nFrames * int(q.fmt.mBytesPerPacket) * int(q.fmt.mFramesPerPacket)
	var err error
	var st C.OSStatus
	for i := 0; i < 3; i++ {
		st, _ = C.AudioQueueAllocateBuffer(q.qRef, C.UInt32(sz), &q.qBufs[i])
		err = caStatus(st)
		if err != nil {
			return err
		}

		slice := (*[1 << 30]byte)(unsafe.Pointer(q.qBufs[i].mAudioData))[:sz]
		for j := 0; j < sz; j++ {
			slice[j] = 0
		}
	}
	return err
}

func (q *aq) initBufs(sz int) {
	c := int(q.fmt.mChannelsPerFrame)
	q.gBufs[0] = Packet{D: make([]float64, sz*c)}
	q.gBufs[1] = Packet{D: make([]float64, sz*c)}
	q.gBufs[2] = Packet{D: make([]float64, sz*c)}
}

func (q *aq) enqueueBufs() {
	for i := 0; i < 3; i++ {
		C.AudioQueueEnqueueBuffer(q.qRef, q.qBufs[i], 0, nil)
	}
}

func (q *aq) initFormat(v snd.Form, codec sample.Codec) {
	initFormat(&q.fmt, v, codec)
	q.Form = v
	q.codec = codec
}

func caStatus(st C.OSStatus) error {
	if st == 0 {
		return nil
	}
	switch st {
	case C.kAudioQueueErr_InvalidParameter:
		return fmt.Errorf("invalid param\n")
	case C.kAudioQueueErr_InvalidDevice:
		return fmt.Errorf("invalid device\n")
	case C.kAudioQueueErr_Permissions:
		return fmt.Errorf("permissions\n")
	case C.kAudioQueueErr_CodecNotFound:
		return fmt.Errorf("codec unsupported\n")
	default:
		return fmt.Errorf("unknown error status: %d\n", st)
	}
}
