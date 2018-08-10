// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build linux
// +build cgo

package sio

import "zikichombo.org/sound/sample"

// #include "alsa/asoundlib.h"
import "C"

var scodec2Alsa = map[sample.Codec]C.snd_pcm_format_t{
	sample.SInt8:     C.SND_PCM_FORMAT_S8,
	sample.SInt16L:   C.SND_PCM_FORMAT_S16_LE,
	sample.SInt16B:   C.SND_PCM_FORMAT_S16_BE,
	sample.SInt24L:   C.SND_PCM_FORMAT_S24_LE,
	sample.SInt24B:   C.SND_PCM_FORMAT_S24_BE,
	sample.SInt32L:   C.SND_PCM_FORMAT_S32_LE,
	sample.SInt32B:   C.SND_PCM_FORMAT_S32_BE,
	sample.SFloat32L: C.SND_PCM_FORMAT_FLOAT_LE,
	sample.SFloat32B: C.SND_PCM_FORMAT_FLOAT_BE,
	sample.SFloat64L: C.SND_PCM_FORMAT_FLOAT64_LE,
	sample.SFloat64B: C.SND_PCM_FORMAT_FLOAT64_BE}

var alsa2sCodec = map[C.snd_pcm_format_t]sample.Codec{
	C.SND_PCM_FORMAT_S8:         sample.SInt8,
	C.SND_PCM_FORMAT_S16_LE:     sample.SInt16L,
	C.SND_PCM_FORMAT_S16_BE:     sample.SInt16B,
	C.SND_PCM_FORMAT_S24_LE:     sample.SInt24L,
	C.SND_PCM_FORMAT_S24_BE:     sample.SInt24B,
	C.SND_PCM_FORMAT_S32_LE:     sample.SInt32L,
	C.SND_PCM_FORMAT_S32_BE:     sample.SInt32B,
	C.SND_PCM_FORMAT_FLOAT_LE:   sample.SFloat32L,
	C.SND_PCM_FORMAT_FLOAT_BE:   sample.SFloat32B,
	C.SND_PCM_FORMAT_FLOAT64_LE: sample.SFloat64L,
	C.SND_PCM_FORMAT_FLOAT64_BE: sample.SFloat64B}
