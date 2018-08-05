// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build linux
// +build cgo

package sio

import (
	"fmt"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// #cgo LDFLAGS: -lasound
// #include "alsa/asoundlib.h"
//
import "C"

//TBD(wsc) figure out how to do this right.
var devices = []*Dev{
	&Dev{Name: "default"},
	&Dev{Name: "plughw:0,0"},
	&Dev{Name: "hw:0,0"}}

func Devices() []*Dev {
	return devices
}

// Input attempts to create and start an Input.
func (d *Dev) Input(v sound.Form, co sample.Codec, n int) (Input, error) {
	pcm := newAlsaPcmIn(d.Name, v, co, n)
	if err := pcm.open(); err != nil {
		return nil, err
	}
	return pcm, nil
}

// Output attempts to create and start an Output, such as to a speaker.
func (d *Dev) Output(v sound.Form, co sample.Codec, n int) (Output, error) {
	pcm := newAlsaPcmOut(d.Name, v, co, n)
	if err := pcm.open(); err != nil {
		return nil, err
	}
	return pcm, nil
}

func init() {
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
	devices = devs[:j]
	DefaultForm = sound.StereoCd()
	DefaultCodec = sample.SFloat32L
	DefaultOutputBufferSize = 256
	DefaultInputBufferSize = 256
}
