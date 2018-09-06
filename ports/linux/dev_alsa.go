// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build linux
// +build cgo

package linux

import (
	"log"
	"time"

	"zikichombo.org/sio/host"
	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

type alsaEntry struct {
	host.NullEntry
}

func (e *alsaEntry) Name() string {
	return "Linux -- ALSACGO"
}

func (e *alsaEntry) DefaultBufferSize() int {
	return 512
}

func (e *alsaEntry) DefaultSampleCodec() sample.Codec {
	return sample.SInt16L
}

func (e *alsaEntry) DefaultForm() sound.Form {
	return sound.StereoCd()
}

func (e *alsaEntry) CanOpenSource() bool {
	return true
}

func (e *alsaEntry) OpenSource(d *libsio.Dev, v sound.Form, co sample.Codec, b int) (sound.Source, time.Time, error) {
	var t time.Time
	pcm := newAlsaPcmIn(d.Name, v, co, b)
	if err := pcm.open(); err != nil {
		return nil, t, err
	}
	return libsio.InputSource(pcm), pcm.pkts[0].Start, nil
}

func (e *alsaEntry) CanOpenSink() bool {
	return true
}

func (e *alsaEntry) OpenSink(d *libsio.Dev, v sound.Form, co sample.Codec, b int) (sound.Sink, *time.Time, error) {
	pcm := newAlsaPcmOut(d.Name, v, co, b)
	if err := pcm.open(); err != nil {
		return nil, nil, err
	}
	return libsio.OutputSink(pcm), &pcm.pkts[0].Start, nil
}

func (e *alsaEntry) HasDevices() bool {
	return true
}

func (e *alsaEntry) Devices() []*libsio.Dev {
	return devices
}

func (e *alsaEntry) DefaultInputDev() *libsio.Dev {
	return devices[0]
}

func (e *alsaEntry) DefaultOutputDev() *libsio.Dev {
	return devices[0]
}

func (e *alsaEntry) DefaultDuplexDev() *libsio.Dev {
	return devices[0]
}

// set up process of ids and free list.
// now, since each id refers to fixed place in memory
// we can use it without locking in the sound data
// processing loop.
func init() {
	e := &alsaEntry{NullEntry: host.NullEntry{}}
	if err := host.RegisterEntry(e); err != nil {
		log.Printf("zc failed load %s: %s\n", e.Name(), err.Error())
	}
}

//TBD(wsc) figure out how to do this right.
var devices = []*libsio.Dev{
	&libsio.Dev{Name: "default"},
	&libsio.Dev{Name: "plughw:0,0"},
	&libsio.Dev{Name: "hw:0,0"}}

// name says it all.
/*
func oldBrokenInit() {
	devs := Devices()
	if len(devs) > 0 {
		DefaultInputDev = devs[0]
		DefaultOutputDev = devs[0]
	}
	j := 0
	for _, d := range devs {
		var pcm *C.snd_pcm_t
		var name = C.CString(d.Name)
		var hwp *C.snd_pcm_hw_params_t
		C.snd_pcm_hw_params_malloc(&hwp)
		defer C.snd_pcm_hw_params_free(hwp)
		defer C.free(unsafe.Pointer(name))
		ret := C.snd_pcm_open(&pcm, name, C.SND_PCM_STREAM_CAPTURE, 0)
		if ret < 0 {
			continue
		}
		C.snd_pcm_hw_params_any(pcm, hwp)
		for _, codec := range sample.Codecs {
			ret = C.snd_pcm_hw_params_test_format(pcm, hwp,
				scodec2Alsa[codec])
			if ret < 0 {
				continue
			}
			d.SampleCodecs = append(d.SampleCodecs, codec)
		}
		fmt.Printf("%s: %v\n", d, d.SampleCodecs)
		devs[j] = d
		j++
		C.snd_pcm_drain(pcm)
		C.snd_pcm_close(pcm)
	}
}
*/
