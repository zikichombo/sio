package darwin

import (
	"fmt"
	"log"
	"time"

	"zikichombo.org/sio/entry"
	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// implements devScanner part of entry.
type devScanner struct {
}

// audio queue services just default to the system default device
// which the user can change.
func (d *devScanner) Devices() []*libsio.Dev {
	return []*libsio.Dev{
		&libsio.Dev{Name: "CoreAudio -- AudioQueueService (on default devices)",
			SampleCodecs: []sample.Codec{
				sample.SInt8, sample.SInt16L, sample.SInt16B, sample.SInt24L,
				sample.SInt24B, sample.SInt32L, sample.SInt32B, sample.SFloat32L,
				sample.SFloat32B}}}
}

// we don't need to distinguish scanning from getting the list, possibly
// cached.
func (d *devScanner) DevScan() []*libsio.Dev {
	return d.Devices()
}

// implements entry.SourceOpener part of entry.Entry.
type inOpener struct {
}

func (i *inOpener) OpenSource(d *libsio.Dev, v sound.Form, co sample.Codec, bsz int) (sound.Source, time.Time, error) {
	var t time.Time
	if !d.SupportsCodec(co) {
		return nil, t, fmt.Errorf("unsupported sample codec: %s\n", co)
	}
	aq, err := newAqin(v, co, bsz)
	if err != nil {
		return nil, t, err
	}
	src := libsio.InputSource(aq)
	return src, aq.gBufs[0].Start, err
}

// implements entry.SinkOpener part of entry.Entry.
type outOpener struct {
}

func (o *outOpener) OpenSink(d *libsio.Dev, v sound.Form, co sample.Codec, bsz int) (sound.Sink, *time.Time, error) {
	if !d.SupportsCodec(co) {
		return nil, nil, fmt.Errorf("unsupported sample codec: %s\n", co)
	}
	aqo, err := newAqo(v, co, bsz)
	if err != nil {
		return nil, nil, err
	}
	snk := libsio.OutputSink(aqo)
	return snk, &aqo.gBufs[0].Start, nil
}

var e = &entry.Entry{
	Name:                    "CoreAudio Audio Queue Services",
	DefaultInputBufferSize:  256,
	DefaultOutputBufferSize: 256,
	DevScanner:              &devScanner{},
	SinkOpener:              &outOpener{},
	SourceOpener:            &inOpener{}}

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
	if err := entry.RegisterEntry(e); err != nil {
		log.Printf("zc failed load %s: %s\n", e.Name, err.Error())
	}
}
