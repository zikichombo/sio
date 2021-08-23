// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package libsio

import (
	"io"

	"github.com/zikichombo/sound"
)

// Input encapsulates input from a device.
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
		return 0, sound.ErrChannelAlignment
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
