// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build darwin
// +build cgo

package sio

import (
	"encoding/binary"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

// #cgo LDFLAGS: -framework AudioToolbox
//
// #include <CoreAudio/CoreAudioTypes.h>
import "C"

func initFormat(fmt *C.AudioStreamBasicDescription, v sound.Form, codec sample.Codec) {
	nCh := C.UInt32(v.Channels())
	fmt.mSampleRate = C.Float64(float64(v.SampleRate()) / float64(freq.Hertz))
	fmt.mFormatID = C.kAudioFormatLinearPCM
	fmt.mChannelsPerFrame = nCh
	fmt.mFramesPerPacket = 1
	fmt.mBytesPerFrame = nCh * C.uint(codec.Bytes())
	fmt.mBytesPerPacket = fmt.mBytesPerFrame * fmt.mFramesPerPacket
	fmt.mBitsPerChannel = C.uint(codec.Bits())
	fmt.mFormatFlags = C.kAudioFormatFlagIsPacked
	if !codec.IsFloat() && codec != sample.SByte {
		fmt.mFormatFlags |= C.kAudioFormatFlagIsSignedInteger
	} else if codec.IsFloat() {
		fmt.mFormatFlags |= C.kAudioFormatFlagIsFloat
	}
	if codec.Bytes() > 1 {
		if codec.ByteOrder() == binary.BigEndian {
			fmt.mFormatFlags |= C.kAudioFormatFlagIsBigEndian
		}
	}
	// NB implicitly interleaved
}
