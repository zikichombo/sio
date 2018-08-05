// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

import "github.com/irifrance/snd"

type Duplex interface {
	snd.Form
	snd.Closer
	InChannels() int
	OutChannels() int

	BeginC() <-chan *DuplexPacket
	EndC() chan<- *DuplexPacket
}
