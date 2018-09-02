// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build darwin
// +build cgo

package darwin

import (
	"errors"
	"fmt"
	"log"
	"math"
	"unsafe"

	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

// #cgo LDFLAGS: -framework CoreServices -framework CoreAudio -framework AudioToolbox
//
// #include <CoreAudio/CoreAudio.h>
// #include <CoreServices/CoreServices.h>
// #include <AudioToolbox/AudioComponent.h>
// #include <AudioToolbox/AUComponent.h>
// #include <AudioToolbox/AudioUnitProperties.h>
import "C"

var devices = []*libsio.Dev{
	&libsio.Dev{Name: "CoreAudio -- AudioQueueService (on default devices)",
		SampleCodecs: []sample.Codec{
			sample.SInt8, sample.SInt16L, sample.SInt16B, sample.SInt24L,
			sample.SInt24B, sample.SInt32L, sample.SInt32B, sample.SFloat32L,
			sample.SFloat32B}}}

func Devices() []*libsio.Dev {
	return devices
}

// SetSampleRate attempts to set the nominal sample rate
// of d to sr.
//
// SetSampleRate returns a non-nil error if the operation
// fails.
//
// The nominal sample rate can be interpreted as the rate
// at which the device tries to operate.
func SetSampleRate(d *libsio.Dev, sr freq.T) error {
	fsr := C.Float64(sr.Float64())
	var pAddr C.AudioObjectPropertyAddress
	pAddr.mScope = C.kAudioDevicePropertyScopeOutput
	pAddr.mElement = 0
	pAddr.mSelector = C.kAudioDevicePropertyNominalSampleRate
	if d.MaxOutChannels == 0 {
		pAddr.mScope = C.kAudioDevicePropertyScopeInput
		pAddr.mElement = 1
	}
	st := C.AudioObjectSetPropertyData(C.AudioObjectID(d.Id), &pAddr, 0, C.NULL,
		C.sizeof_Float64, unsafe.Pointer(&fsr))
	return caStatus(st)
}

func SetBufferSize(d *libsio.Dev, sz int) error {
	csz := C.UInt32(sz)
	var pAddr C.AudioObjectPropertyAddress
	pAddr.mSelector = C.kAudioDevicePropertyBufferFrameSize
	if d.MaxOutChannels > 0 {
		pAddr.mScope = C.kAudioDevicePropertyScopeInput
		pAddr.mElement = 0
		st := C.AudioObjectSetPropertyData(C.AudioObjectID(d.Id), &pAddr, 0, C.NULL,
			C.sizeof_UInt32, unsafe.Pointer(&csz))
		if err := caStatus(st); err != nil {
			return err
		}
	}
	if d.MaxInChannels > 0 {
		pAddr.mScope = C.kAudioDevicePropertyScopeOutput
		pAddr.mElement = 1
		st := C.AudioObjectSetPropertyData(C.AudioObjectID(d.Id), &pAddr, 0, C.NULL,
			C.sizeof_UInt32, unsafe.Pointer(&csz))
		if err := caStatus(st); err != nil {
			return err
		}
	}
	return nil
}

// Input attempts to create and start an Input.
func mkInput(d *libsio.Dev, v sound.Form, co sample.Codec, n int) (libsio.Input, error) {
	if !d.SupportsCodec(co) {
		return nil, fmt.Errorf("unsupported sample codec: %s\n", co)
	}
	return newAqin(v, co, n)
}

// Output attempts to create and start an Output, such as to a speaker.
func mkOutput(d *libsio.Dev, v sound.Form, co sample.Codec, n int) (libsio.Output, error) {
	if !d.SupportsCodec(co) {
		return nil, fmt.Errorf("unsupported sample codec: %s\n", co)
	}
	return newAqo(v, co, n)
}

func list() error {
	var sys C.AudioObjectID
	sys = C.kAudioObjectSystemObject
	var sz C.UInt32
	var devAddr C.AudioObjectPropertyAddress
	devAddr.mSelector = C.kAudioHardwarePropertyDevices
	devAddr.mScope = C.kAudioObjectPropertyScopeGlobal
	devAddr.mElement = 0
	st := C.AudioObjectGetPropertyDataSize(sys, &devAddr, 0, C.NULL, &sz)
	if err := caStatus(st); err != nil {
		return err
	}
	nDevs := int(sz / C.sizeof_AudioObjectID)
	_ = nDevs
	fmt.Printf("list() %d devs\n", nDevs)
	devBuf := C.malloc(C.ulong(sz))
	defer C.free(devBuf)
	st = C.AudioObjectGetPropertyData(sys, &devAddr, 0, C.NULL, &sz, devBuf)
	if err := caStatus(st); err != nil {
		return err
	}
	devs := (*[1 << 30]C.AudioObjectID)(unsafe.Pointer(devBuf))[:nDevs]
	fmt.Printf("devs %v\n", devs)
	dIn, dOut, dSys, err := hwDefaults()
	if err != nil {
		return err
	}
	Devs := make([]libsio.Dev, 0, 7)

	for _, dev := range devs {
		N := len(Devs)
		Devs = append(Devs, libsio.Dev{Id: uint64(dev)})
		D := &Devs[N]
		nm, err := name(dev)
		if err != nil {
			log.Printf("error getting name of device %d: %s\n", dev, err)
			continue
		}
		D.Name = nm

		ic, err := channels(dev, C.kAudioDevicePropertyScopeInput)
		if err != nil {
			log.Printf("error getting input channels of device %d: %s\n", dev, err)
			continue
		}
		D.MaxInChannels = ic
		oc, err := channels(dev, C.kAudioDevicePropertyScopeOutput)
		if err != nil {
			log.Printf("error getting output channels of device %d: %s\n", dev, err)
			continue
		}
		D.MaxOutChannels = oc
		if dev == dIn {
			D.IsDefaultIn = true
		}
		if dev == dOut {
			D.IsDefaultOut = true
		}
		if dev == dSys {
			D.IsDefaultSys = true
			// TBD: DefaultSysDev
		}
		if D.MaxOutChannels+D.MaxInChannels == 0 {
			continue
		}
		var minSr, maxSr freq.T
		if D.MaxOutChannels > 0 {
			minSr, maxSr, err = devSrs(dev, C.kAudioDevicePropertyScopeOutput)
			if err != nil {
				return err
			}
		} else if D.MaxInChannels > 0 {
			minSr, maxSr, err = devSrs(dev, C.kAudioDevicePropertyScopeInput)
			if err != nil {
				return err
			}
		}
		D.MinSampleRate = minSr
		D.MaxSampleRate = maxSr
		fmt.Printf("%s\n", D)
	}
	return nil
}

func devSrs(dev C.AudioObjectID, scope C.AudioObjectPropertyScope) (freq.T, freq.T, error) {
	var pAddr C.AudioObjectPropertyAddress
	pAddr.mScope = scope
	if scope == C.kAudioDevicePropertyScopeOutput {
		pAddr.mElement = 0
	} else {
		pAddr.mElement = 1
	}
	pAddr.mSelector = C.kAudioDevicePropertyAvailableNominalSampleRates
	var sz C.UInt32
	st := C.AudioObjectGetPropertyDataSize(dev, &pAddr, 0, C.NULL, &sz)
	if err := caStatus(st); err != nil {
		return 0, 0, err
	}
	rangesPtr := (*C.AudioValueRange)(C.malloc(C.ulong(sz)))
	defer C.free(unsafe.Pointer(rangesPtr))
	st = C.AudioObjectGetPropertyData(dev, &pAddr, 0, C.NULL, &sz, unsafe.Pointer(rangesPtr))
	if err := caStatus(st); err != nil {
		return 0, 0, err
	}
	nRanges := sz / C.sizeof_AudioValueRange
	ranges := (*[1 << 30]C.AudioValueRange)(unsafe.Pointer(rangesPtr))[:nRanges]
	min := 1e10
	max := 1.0
	for i := range ranges {
		r := &ranges[i]
		if float64(r.mMinimum) < min {
			min = float64(r.mMinimum)
		}
		if float64(r.mMaximum) > max {
			max = float64(r.mMaximum)
		}
	}
	minF := freq.T(uint64(math.Floor(min+0.5))) * freq.Hertz
	maxF := freq.T(uint64(math.Floor(max+0.5))) * freq.Hertz
	if minF > maxF {
		return 0, 0, errors.New("disjoint ranges not supported")
	}
	return minF, maxF, nil
}

func hwDefaults() (C.AudioObjectID, C.AudioObjectID, C.AudioObjectID, error) {
	var sys C.AudioObjectID
	sys = C.kAudioObjectSystemObject
	sz := C.uint(C.sizeof_AudioObjectID)
	var dIn, dOut, dSys C.AudioObjectID

	var pAddr C.AudioObjectPropertyAddress
	pAddr.mScope = C.kAudioObjectPropertyScopeGlobal
	pAddr.mElement = 0
	pAddr.mSelector = C.kAudioHardwarePropertyDefaultInputDevice
	st := C.AudioObjectGetPropertyData(sys, &pAddr, 0, C.NULL, &sz, unsafe.Pointer(&dIn))
	if err := caStatus(st); err != nil {
		return 0, 0, 0, err
	}
	pAddr.mSelector = C.kAudioHardwarePropertyDefaultOutputDevice
	st = C.AudioObjectGetPropertyData(sys, &pAddr, 0, C.NULL, &sz, unsafe.Pointer(&dOut))
	if err := caStatus(st); err != nil {
		return 0, 0, 0, err
	}
	pAddr.mSelector = C.kAudioHardwarePropertyDefaultSystemOutputDevice
	st = C.AudioObjectGetPropertyData(sys, &pAddr, 0, C.NULL, &sz, unsafe.Pointer(&dSys))
	if err := caStatus(st); err != nil {
		return 0, 0, 0, err
	}
	return dIn, dOut, dSys, nil
}

func channels(dev C.AudioObjectID, scope C.AudioObjectPropertyScope) (int, error) {
	var pAddr C.AudioObjectPropertyAddress
	pAddr.mSelector = C.kAudioDevicePropertyStreamConfiguration
	pAddr.mScope = scope
	pAddr.mElement = 0
	var sz C.UInt32
	res := C.AudioObjectGetPropertyDataSize(dev, &pAddr, 0, C.NULL, &sz)
	if err := caStatus(res); err != nil {
		return 0, err
	}
	bufferList := (*C.AudioBufferList)(C.malloc(C.ulong(sz)))
	defer C.free(unsafe.Pointer(bufferList))
	res = C.AudioObjectGetPropertyData(dev, &pAddr, 0, C.NULL, &sz, unsafe.Pointer(bufferList))
	if err := caStatus(res); err != nil {
		return 0, err
	}
	nStr := bufferList.mNumberBuffers
	bufs := (*[1 << 30]C.AudioBuffer)(unsafe.Pointer(&bufferList.mBuffers))[:nStr]
	nC := 0
	for i := range bufs {
		buf := &bufs[i]
		nC += int(buf.mNumberChannels)
	}
	return nC, nil
}

func name(dev C.AudioObjectID) (string, error) {
	return strProperty(dev, C.kAudioObjectPropertyName)
}

func strProperty(dev C.AudioObjectID, sel C.AudioObjectPropertySelector) (string, error) {
	var devAddr C.AudioObjectPropertyAddress
	devAddr.mSelector = sel
	devAddr.mScope = C.kAudioObjectPropertyScopeGlobal
	devAddr.mElement = C.kAudioObjectPropertyElementMaster

	var cfs C.CFStringRef
	cfsSz := C.uint(C.sizeof_CFStringRef)
	st := C.AudioObjectGetPropertyData(dev, &devAddr, 0, C.NULL, &cfsSz, unsafe.Pointer(&cfs))
	if err := caStatus(st); err != nil {
		return "", err
	}
	length := C.ulong(2*C.CFStringGetLength(cfs) + 1)
	var cName *C.char
	cName = (*C.char)(C.malloc(length))
	defer C.free(unsafe.Pointer(cName))
	C.CFStringGetCString(cfs, cName, C.long(length), C.kCFStringEncodingUTF8)
	return C.GoString(cName), nil
}

func init() {
	list()
	devs := Devices()
	if len(devs) > 0 {
	}
}
