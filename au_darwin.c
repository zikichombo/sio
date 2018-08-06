#ifndef _SIO_AUHAL_DARWIN
#define _SIO_AUHAL_DARWIN

#include <AudioToolbox/AudioUnit.h>
#include <CoreAudio/CoreAudio.h>

#include <stdatomic.h>

typedef struct HalThunk {
	AudioUnit au;
	AudioBufferList ioBufs[2];
} HalThunk;

void inputHalCallback(
		void *clientData,
		AudioUnitRenderActionFlags *flags,
		const AudioTimeStamp *inTimeStamp,
		UInt32 inBus,
		UInt32 inFrames,
		AudioBufferList *unusedIoData) {

	HalThunk *thunk = (HalThunk *) clientData;
	OSStatus st = AudioUnitRender(thunk->au, flags, inTimeStamp, inBus, inFrames, &thunk->ioBufs[1]);
	if (st != 0) {
		return;
	}
	// notify 
}

void outputHalCallback(
		void *clientData,
		AudioUnitRenderActionFlags *flags,
		const AudioTimeStamp *inTimeStamp,
		UInt32 inBus,
		UInt32 inFrames,
		AudioBufferList *ioData)
{
	HalThunk *thunk = (HalThunk *) clientData;
	// if duplex, render

	// notify 
}

#endif
