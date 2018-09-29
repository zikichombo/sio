// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

#ifndef ZC_SIO_AU_DARWIN_H
#define ZC_SIO_AU_DARWIN_H

#include <AudioToolbox/AudioUnit.h>
#include <CoreAudio/CoreAudio.h>

#include "../../libsio/cb_extern.h" // in sio/libsio, CFLAGS set in au_darwin.go import "C" preamble

typedef struct HalThunk {
	AudioUnit au;
	AudioBufferList ioBuf;
	Cb *cb;
	AURenderCallbackStruct aucb;
} HalThunk;

HalThunk * newHalThunk(AudioUnit au, Cb *cb, int nc, int nf, int bps);
void freeHalThunk(HalThunk *t);
OSStatus setCbIn(HalThunk *t);
OSStatus setCbOut(HalThunk *t);
OSStatus setCbDuplex(HalThunk *t);


#endif
