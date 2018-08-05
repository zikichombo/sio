// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

import (
	"io"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// Interface Input encapsulates input from a device.
type Input interface {
	sound.Form
	sound.Closer

	// C returns a channel from which buffers of samples can be received.  The
	// sample buffers are interleaved.
	//
	// Sample buffers are owned by the object implementing Input.  A client
	// should only use one buffer between each receive from the channel.
	// Otherwise, race conditions can occur.  The object implementing Input
	// must do so in a way that it can fill the next sample buffer, or buffers,
	// while the client is processing the previous one.
	C() <-chan *Packet
}

// NewInput tries to open and start an input device.
// v Gives the valve information, c the dataformat of individual samples,
// and n the buffer size, in frames, of each packet.
func NewInput(v sound.Form, c sample.Codec, n int) (Input, error) {
	return DefaultInputDev.Input(v, c, n)
}

// DefaultInput tries to open and start an input
// device with default sampling rate, number of channels and
// buffersize.
func DefaultInput() (Input, error) {
	return NewInput(DefaultForm, DefaultCodec, DefaultInputBufferSize)
}

func Record() (sound.Source, error) {
	i, e := DefaultInput()
	if e != nil {
		return nil, e
	}
	return InputSource(i), nil
}

func RecordWith(v sound.Form, c sample.Codec, n int) (sound.Source, error) {
	i, e := NewInput(v, c, n)
	if e != nil {
		return nil, e
	}
	return InputSource(i), nil
}

type chn struct {
	sound.Form
	in  Input
	ch  <-chan *Packet
	buf []float64
	p   int
}

func (ch *chn) Close() error {
	ch.in.Close()
	return nil
}

func (ch *chn) Receive(dst []float64) (int, error) {
	nC := ch.Channels()
	if len(dst)%nC != 0 {
		panic("wilma")
		return 0, sound.ChannelAlignmentError
	}
	nF := len(dst) / nC
	var f, c int
	for f < nF {
		if ch.p == len(ch.buf) {
			pkt, open := <-ch.ch
			if !open {
				if f == 0 {
					return 0, io.EOF
				}
				return f, nil
			}
			ch.buf = pkt.D
			ch.p = 0
		}
		dst[c*nF+f] = ch.buf[ch.p]
		ch.p++
		c++
		if c == nC {
			c = 0
			f++
		}
	}
	return f, nil
}

// InputSource returns a source from an input.
func InputSource(in Input) sound.Source {
	return &chn{
		Form: in,
		in:   in,
		ch:   in.C()}
}
