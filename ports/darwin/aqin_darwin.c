#ifndef _AUDIO_INAQ_DARWIN
#define _AUDIO_INAQ_DARWIN

#include <AudioToolbox/AudioQueue.h>
#include <CoreAudio/CoreAudioTypes.h>
#include <CoreFoundation/CFRunLoop.h>


void inGoCallback(int, AudioQueueBufferRef, char *, size_t);

void inputCallback(
		void *goInput, 
		AudioQueueRef q, 
		AudioQueueBufferRef buf,
		const AudioTimeStamp* unusedTS,
		UInt32 unusedNumPackets,
		const AudioStreamPacketDescription* unusedPD
		) {
  inGoCallback((int)(goInput), buf, buf->mAudioData, buf->mAudioDataByteSize);
}

#endif
