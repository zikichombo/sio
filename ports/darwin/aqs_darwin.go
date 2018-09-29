// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package darwin

import (
	"log"
	"time"

	"zikichombo.org/sio/host"
	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

type aqsEntry struct {
	host.NullEntry
}

func (e *aqsEntry) Name() string {
	return "CoreAudio Audio Queue Services"
}

func (e *aqsEntry) DefaultBufferSize() int {
	return 256
}

func (e *aqsEntry) DefaultSampleCodec() sample.Codec {
	return sample.SFloat32L
}

func (e *aqsEntry) DefaultForm() sound.Form {
	return sound.MonoCd()
}

func (e *aqsEntry) CanOpenSource() bool {
	return true
}

func (e *aqsEntry) OpenSource(d *libsio.Dev, v sound.Form, co sample.Codec, b int) (sound.Source, time.Time, error) {
	var t time.Time
	aq, err := newAqin(v, co, b)
	if err != nil {
		return nil, t, err
	}
	src := libsio.InputSource(aq)
	return src, aq.gBufs[0].Start, err
}

func (e *aqsEntry) CanOpenSink() bool {
	return true
}

func (e *aqsEntry) OpenSink(d *libsio.Dev, v sound.Form, co sample.Codec, b int) (sound.Sink, *time.Time, error) {
	aqo, err := newAqo(v, co, b)
	if err != nil {
		return nil, nil, err
	}
	snk := libsio.OutputSink(aqo)
	return snk, &aqo.gBufs[0].Start, nil
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
	e := &aqsEntry{NullEntry: host.NullEntry{}}
	if err := host.RegisterEntry(e); err != nil {
		log.Printf("zc failed load %s: %s\n", e.Name(), err.Error())
	}
}
