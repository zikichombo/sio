// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

#include "au_darwin.h"

HalThunk * newHalThunk(AudioUnit au, Cb *cb, int nc, int nf, int bps) {
	HalThunk * ht = (HalThunk *) malloc(sizeof(HalThunk));
	if (ht == 0) {
		return ht;
	}
	ht->au = au;
	ht->cb = cb;
	ht->aucb.inputProcRefCon = ht;
	ht->ioBuf.mNumberBuffers = 1;
	ht->ioBuf.mBuffers[0].mNumberChannels = nc;
	ht->ioBuf.mBuffers[0].mDataByteSize = nc*nf*bps;
	ht->ioBuf.mBuffers[0].mData = calloc(bps*nc, nf);
	if (ht->ioBuf.mBuffers[0].mData == 0) {
		free(ht);
		return 0;
	}
	return ht;
}

void freeHalThunk(HalThunk *t) {
	free(t->ioBuf.mBuffers[0].mData);
	free(t);
}

OSStatus inputHalCallback(
		void *clientData,
		AudioUnitRenderActionFlags *flags,
		const AudioTimeStamp *inTimeStamp,
		UInt32 inBus,
		UInt32 inFrames,
		AudioBufferList *unusedIoData) {

	HalThunk *thunk = (HalThunk *) clientData;
	OSStatus st = AudioUnitRender(thunk->au, flags, inTimeStamp, inBus, inFrames, &thunk->ioBuf);
	if (st != 0) {
		return st;
	}
	Cb *cb = thunk->cb;
	cb->inCb(cb, thunk->ioBuf.mBuffers[0].mData, (int)inFrames);
	return 0;
}

OSStatus outputHalCallback(
		void *clientData,
		AudioUnitRenderActionFlags *flags,
		const AudioTimeStamp *inTimeStamp,
		UInt32 inBus,
		UInt32 inFrames,
		AudioBufferList *ioData)
{
	HalThunk *thunk = (HalThunk *) clientData;
	// NB we have set the frames by construction, so they should not vary
	// w.r.t. thunk->cb expected frame size, thus we don't need to pay
	// attention to the value of inf after the call.
	int of = (int)inFrames;
	Cb *cb = thunk->cb;
	cb->outCb(cb, thunk->ioBuf.mBuffers[0].mData, &of);
	return 0;
}

OSStatus duplexHalCallback(
		void *clientData,
		AudioUnitRenderActionFlags *flags,
		const AudioTimeStamp *inTimeStamp,
		UInt32 inBus,
		UInt32 inFrames,
		AudioBufferList *ioData)
{
	HalThunk *thunk = (HalThunk *) clientData;
	OSStatus st = AudioUnitRender(thunk->au, flags, inTimeStamp, inBus, inFrames, &thunk->ioBuf);
	if (st != 0) {
		return st;
	}
	// NB we have set the frames by construction, so they should not vary
	// w.r.t. thunk->cb expected frame size, thus we don't need to pay
	// attention to the value of inf after the call.
	int inf = (int)inFrames;
	Cb *cb = thunk->cb;
	cb->duplexCb(cb, ioData->mBuffers[0].mData, &inf, thunk->ioBuf.mBuffers[0].mData, inf);
	return 0;
}

OSStatus setCbIn(HalThunk *t) {
	t->aucb.inputProc = inputHalCallback;
	OSStatus st = AudioUnitSetProperty(t->au,
			kAudioOutputUnitProperty_SetInputCallback,
			kAudioUnitScope_Global,
			0,
			&t->aucb,
			sizeof(t->aucb));
	return st;
}

OSStatus setCbOut(HalThunk *t) {
	t->aucb.inputProc = outputHalCallback;
	OSStatus st = AudioUnitSetProperty(t->au,
			kAudioUnitProperty_SetRenderCallback,
			kAudioUnitScope_Input,
			0,
			&t->aucb,
			sizeof(t->aucb));
	return st;
}

OSStatus setCbDuplex(HalThunk *t) {
	t->aucb.inputProc = duplexHalCallback;
	OSStatus st = AudioUnitSetProperty(t->au,
			kAudioUnitProperty_SetRenderCallback,
			kAudioUnitScope_Input,
			0,
			&t->aucb,
			sizeof(t->aucb));
	return 0;
}

