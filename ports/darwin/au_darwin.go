// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

//go:build darwin && cgo
// +build darwin,cgo

package darwin

import (
	"fmt"
	"log"
	"math"
	"unsafe"

	"github.com/zikichombo/sio/libsio"
	"github.com/zikichombo/sound"
	"github.com/zikichombo/sound/sample"
)

// #cgo LDFLAGS: -framework CoreServices -framework CoreAudio -framework AudioToolbox
//
// #include "../../libsio/cb.h"
// #include "au_darwin.h"
//
// #include <CoreAudio/CoreAudio.h>
// #include <CoreServices/CoreServices.h>
// #include <AudioToolbox/AudioComponent.h>
// #include <AudioToolbox/AUComponent.h>
// #include <AudioToolbox/AudioUnitProperties.h>
// #include <AudioToolbox/AudioOutputUnit.h>
import "C"

type auhal struct {
	*libsio.Cb
	u       C.AudioUnit
	thunk   *C.HalThunk
	dev     *libsio.Dev
	iom     libsio.IoMode
	form    sound.Form
	codec   sample.Codec
	bufSz   int
	started bool
}

func newAuio(dev *libsio.Dev, iom libsio.IoMode, v sound.Form, co sample.Codec, bufSz int) (*auhal, error) {
	if err := setSampleRate(dev, v.SampleRate()); err != nil {
		return nil, err
	}
	var ud C.AudioComponentDescription
	ud.componentType = C.kAudioUnitType_Output
	ud.componentSubType = C.kAudioUnitSubType_HALOutput // on iOS kAudioUnitSubType_RemoteIO
	ud.componentManufacturer = C.kAudioUnitManufacturer_Apple
	ud.componentFlags = 0
	ud.componentFlagsMask = 0

	var ac C.AudioComponent
	ac = C.AudioComponentFindNext((C.AudioComponent)(C.NULL), &ud)
	var u C.AudioComponentInstance
	st := C.AudioComponentInstanceNew(ac, &u)
	if err := caStatus(st); err != nil {
		log.Printf("error instantiating I/O unit: %s\n", err)
		return nil, err
	}
	res := &auhal{Cb: libsio.NewCb(v, co, bufSz), u: u, iom: iom, form: v, codec: co}
	if err := res.enableIO(); err != nil {
		return nil, err
	}
	if err := res.setDev(dev); err != nil {
		return nil, err
	}
	if err := res.setFormats(v, co); err != nil {
		return nil, err
	}
	if err := res.allocBufs(bufSz); err != nil {
		return nil, err
	}
	res.thunk = C.newHalThunk(res.u, (*C.Cb)(res.Cb.C()), C.int(v.Channels()), C.int(bufSz), C.int(co.Bytes()))
	if res.thunk == nil {
		return nil, fmt.Errorf("oom")
	}
	if err := res.setCb(); err != nil {
		return nil, err
	}

	st = C.AudioUnitInitialize(u)
	if err := caStatus(st); err != nil {
		log.Printf("error instantiating I/O unit: %s\n", err)
		return nil, err
	}
	return res, nil
}

func (u *auhal) enableIO() error {
	x := C.int(1)
	if !u.iom.Inputs() {
		x = 0
	}
	st := C.AudioUnitSetProperty(u.u, C.kAudioOutputUnitProperty_EnableIO,
		C.kAudioUnitScope_Input,
		1, // input element
		unsafe.Pointer(&x),
		C.sizeof_int)
	if err := caStatus(st); err != nil {
		log.Printf("error enabling input: %s\n", err)
		return err
	}
	x = C.int(1)
	if !u.iom.Outputs() {
		x = 0
	}
	st = C.AudioUnitSetProperty(u.u, C.kAudioOutputUnitProperty_EnableIO,
		C.kAudioUnitScope_Output,
		0, // output element
		unsafe.Pointer(&x),
		C.sizeof_int)
	if err := caStatus(st); err != nil {
		log.Printf("error enabling output: %s\n", err)
		return err
	}
	return nil
}

func (u *auhal) setDev(dev *libsio.Dev) error {
	iom := u.iom
	if dev.MaxOutChannels == 0 && iom.Outputs() {
		return fmt.Errorf("no output channels on device %s\n", dev.Name)
	}
	if dev.MaxInChannels == 0 && iom.Inputs() {
		return fmt.Errorf("no output channels on device %s\n", dev.Name)
	}
	if err := setBufferSize(dev, u.bufSz); err != nil {
		return err
	}
	id := C.AudioObjectID(u.dev.Id)
	st := C.AudioUnitSetProperty(u.u, C.kAudioOutputUnitProperty_CurrentDevice,
		C.kAudioUnitScope_Global, 0, unsafe.Pointer(&id), C.sizeof_AudioObjectID)
	return caStatus(st)
}

func (u *auhal) setFormats(v sound.Form, co sample.Codec) error {
	var devFmt, appFmt C.AudioStreamBasicDescription
	sz := C.uint(C.sizeof_AudioStreamBasicDescription)
	initFormat(&appFmt, v, co)
	if u.iom.Inputs() {
		st := C.AudioUnitGetProperty(u.u, C.kAudioUnitProperty_StreamFormat,
			C.kAudioUnitScope_Input, 1, unsafe.Pointer(&devFmt), &sz)
		if err := caStatus(st); err != nil {
			return err
		}
		if math.Abs(float64(devFmt.mSampleRate)-float64(appFmt.mSampleRate)) > 1.0 {
			return fmt.Errorf("unable to obtain requested sample rate")
		}
		appFmt.mSampleRate = devFmt.mSampleRate
		st = C.AudioUnitSetProperty(u.u, C.kAudioUnitProperty_StreamFormat,
			C.kAudioUnitScope_Input, 0, unsafe.Pointer(&appFmt), sz)
		if err := caStatus(st); err != nil {
			return err
		}
	}
	if u.iom.Outputs() {
		st := C.AudioUnitGetProperty(u.u, C.kAudioUnitProperty_StreamFormat,
			C.kAudioUnitScope_Output, 0, unsafe.Pointer(&devFmt), &sz)
		if err := caStatus(st); err != nil {
			return err
		}
		if math.Abs(float64(devFmt.mSampleRate)-float64(appFmt.mSampleRate)) > 1.0 {
			return fmt.Errorf("unable to obtain requested sample rate")
		}
		appFmt.mSampleRate = devFmt.mSampleRate
		st = C.AudioUnitSetProperty(u.u, C.kAudioUnitProperty_StreamFormat,
			C.kAudioUnitScope_Output, 1, unsafe.Pointer(&appFmt), sz)
		if err := caStatus(st); err != nil {
			return err
		}
	}
	return nil
}

func (u *auhal) allocBufs(bufSz int) error {
	sz := C.uint(C.sizeof_UInt32)
	var frames = C.UInt32(bufSz)
	if u.iom.Outputs() {
		st := C.AudioUnitSetProperty(u.u, C.kAudioDevicePropertyBufferFrameSize,
			C.kAudioUnitScope_Output, 0, unsafe.Pointer(&frames), sz)
		if err := caStatus(st); err != nil {
			return err
		}
	}
	if u.iom.Inputs() {
		st := C.AudioUnitSetProperty(u.u, C.kAudioDevicePropertyBufferFrameSize,
			C.kAudioUnitScope_Input, 1, unsafe.Pointer(&frames), sz)
		if err := caStatus(st); err != nil {
			return err
		}
	}
	return nil
}

func (u *auhal) setCb() error {
	switch u.iom {
	case libsio.InputMode:
		st := C.setCbIn(u.thunk)
		if err := caStatus(st); err != nil {
			return err
		}
	case libsio.OutputMode:
		st := C.setCbOut(u.thunk)
		if err := caStatus(st); err != nil {
			return err
		}
	case libsio.DuplexMode:
		st := C.setCbDuplex(u.thunk)
		if err := caStatus(st); err != nil {
			return err
		}
	default:
		panic("unreachable")
	}
	return nil
}

func (u *auhal) Close() error {
	st := C.AudioOutputUnitStop(u.u)
	if err := caStatus(st); err != nil {
		return err
	}
	st = C.AudioUnitUninitialize(u.u)
	if err := caStatus(st); err != nil {
		return err
	}
	st = C.AudioComponentInstanceDispose(u.u)
	if err := caStatus(st); err != nil {
		return err
	}
	C.freeHalThunk(u.thunk)
	return nil
}

func (u *auhal) start() error {
	if u.started {
		return nil
	}
	st := C.AudioOutputUnitStart(u.u)
	err := caStatus(st)
	if err != nil {
		u.started = true
	}
	return err
}

func (u *auhal) Send(d []float64) error {
	if err := u.start(); err != nil {
		return err
	}
	return u.Cb.Send(d)
}

func (u *auhal) Receive(d []float64) (int, error) {
	if err := u.start(); err != nil {
		return 0, err
	}
	return u.Cb.Receive(d)
}
