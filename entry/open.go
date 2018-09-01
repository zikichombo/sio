package entry

import (
	"time"

	"zikichombo.org/sio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// SourceOpener is an interface for opening a sound.Source from input
// such as a microphone.
type SourceOpener interface {
	// OpenSource starts capturing audio.
	//
	// dev is the specified device, and should be nil if and only if the current
	// entry has .DevScanner nil.  Otherwise, dev should be a device returned by
	// the current entry .DevScan()
	//
	// v indicates the desired form (channels, sample rate) of the source.
	//
	// sco indicates the desired sample.Codec.
	//
	// bufSz indicates the size of buffer whose data is placed in
	// sound.Source.Receive.  This is normally the size of the part of the ring
	// buffer exposed for reading.  Implementations should use a minimal total
	// buffer size to accomodate this constraint.  and capture to the non-exposed
	// part of the buffer while the caller receives from the buffer.
	//
	// OpenSource returns a triple (s, t, e) with
	// s: sound.Source which represents captured audio.
	// t: the start time of the first sample.
	// e: any error
	OpenSource(dev *sio.Dev, v sound.Form, sco sample.Codec, bufSz int) (sound.Source, time.Time, error)
}

// SinkOpener is an interface for opening a sound.Sink to output such as
// a speaker.
type SinkOpener interface {
	// OpenSink starts playing audio.
	//
	// dev is the specified device, and should be nil if and only if the current
	// entry has .DevScanner nil.  Otherwise, dev should be a device returned by
	// the current entry .DevScan()
	//
	// v indicates the desired form (channels, sample rate) of the source.
	//
	// sco indicates the desired sample.Codec.
	//
	// bufSz indicates the size of buffer whose data is placed
	// in sound.Source.Receive.  This is normally the size of the part of
	// the ring buffer exposed for playback.  Implementations should
	// use a minimal total buffer size to accomodate this constraint
	// and
	// capture to the non-exposed part of the buffer while the caller
	// receives from the buffer.
	//
	// OpenSink returns a tuple (s, *t, e) with
	// s: sound.Sink which represents audio for playback.
	// t: a pointer to the start time of the first played sample, which is set
	//    after the first successful send via the returned sound.Sink.
	// e: any error
	//
	// OpenSink should
	OpenSink(dev *sio.Dev, v sound.Form, sco sample.Codec, bufSz int) (sound.Sink, *time.Time, error)
}

// DuplexOpener is an interface for opening Duplex connections.
type DuplexOpener interface {
	// OpenDuplex
	OpenDuplex(dev *sio.Dev, v sound.Form, sco sample.Codec, bufSz int) (sound.Duplex, time.Time, error)
}
