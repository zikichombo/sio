#ifndef _AUDIO_AQO_DARWIN
#define _AUDIO_AQO_DARWIN

#include <AudioToolbox/AudioQueue.h>
#include <CoreAudio/CoreAudioTypes.h>
#include <CoreFoundation/CFRunLoop.h>


void outGoCallback(int, AudioQueueBufferRef, char *, size_t);

void outputCallback(
		void *goOutput, 
		AudioQueueRef q, 
		AudioQueueBufferRef buf) {
	outGoCallback((int)goOutput, buf, buf->mAudioData, buf->mAudioDataBytesCapacity);
}


#endif
