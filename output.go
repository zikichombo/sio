// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

import (
	"errors"

	"zikichombo.org/sound"
	"zikichombo.org/sound/ops"
	"zikichombo.org/sound/sample"
)

// Output encapsulates an output device such as to a speaker.
type Output interface {
	sound.Form
	sound.Closer

	// FillC returns a channel for receiving buffers to be filled and subsequently
	// sent on PlayC() to output to the final destination.
	//
	// If Output.Close() is called, or if PlayC() is closed, then FillC may be closed
	// without sending a buffer.  The channel should not be used as a source of buffers
	// more than once between each corresponding send on PlayC().
	FillC() <-chan *Packet

	// PlayC accepts buffers originating from FillC and adds the data to the queue
	// of data to be played.  Output may be closed by closing PlayC() or by calling
	// Output.Close().
	//
	// If the memory backing the sent slice is not the same as that previously sent
	// on FillC(), then PlayC() panics.
	//
	// A send on PlayC() should occur at most once between receives from FillC().
	//
	// Playback may be scheduled in the future by adding to packet.N.
	// once a packet has been sent on the playback channel with packet.N
	// set, all subsequent packets from the FillC() have this offset
	// incorporated in their Packet.N values.  If a system does not support
	// this functionality, it should close the device and log an error
	// when the client attempts to schedule in the future.
	PlayC() chan<- *Packet
}

// NewOutput attempts to open and start an output device.
// v Gives the form (channels and sample rate), c the dataformat of individual samples,
// and n the buffer size, in frames, of each packet.
func NewOutput(v sound.Form, c sample.Codec, n int) (Output, error) {
	return DefaultOutputDev.Output(v, c, n)
}

// DefaultOutput attempts to open an output stream with default
// valve, sample codec, and output buffer size.
func DefaultOutput() (Output, error) {
	return NewOutput(DefaultForm, DefaultCodec, DefaultOutputBufferSize)
}

type osnk struct {
	sound.Form
	out   Output
	fillC <-chan *Packet
	playC chan<- *Packet
	pkt   *Packet
	p     int
}

func (o *osnk) Close() error {
	o.out.Close()
	return nil
}

func (o *osnk) Send(d []float64) error {
	nC := o.Channels()
	if len(d)%nC != 0 {
		return sound.ChannelAlignmentError
	}
	nF := len(d) / nC
	var c, f int
	for f < nF {
		if o.pkt == nil {
			pkt, ok := <-o.fillC
			if !ok {
				return errors.New("output closed")
			}
			o.pkt = pkt
		}
		o.pkt.D[o.p] = d[c*nF+f]
		o.p++
		if o.p == len(o.pkt.D) {
			o.playC <- o.pkt
			o.p = 0
			o.pkt = nil
		}
		c++
		if c == nC {
			c = 0
			f++
		}
	}
	return nil
}

// OutputSink converts an output to a Sink.
func OutputSink(o Output) sound.Sink {
	return &osnk{
		Form:  o,
		out:   o,
		fillC: o.FillC(),
		playC: o.PlayC()}
}

// Play plays the Source src using the defaults.
//
// Play returns a non-nil error in case any problems
// occured either creating an output from a device or
// in the actual playback.
func Play(src sound.Source) error {
	return PlaySource(src, DefaultOutputDev.SampleCodecs[0], DefaultOutputBufferSize)
}

// PlaySource is like Play, but allows setting the sample
// codec and packet buffer size.
//
// On some systems (such as alsa), the packet buffer size is only a hint.
func PlaySource(src sound.Source, cod sample.Codec, bufSz int) error {
	o, err := NewOutput(src, cod, bufSz)
	if err != nil {
		return err
	}
	snk := OutputSink(o)
	defer snk.Close()
	return ops.Copy(snk, src)
}
