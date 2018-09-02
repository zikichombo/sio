// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package libsio

import "time"

// Packet represents data which is exchanged between software and a device.
type Packet struct {
	D     []float64 // slice of data, channel-interleaved.
	N     int       // frame number of first element.  For playback, this can be used to schedule in the future.
	Start time.Time // time of first sample in stream.  This is approximate but normally very close.
}

type DuplexPacket struct {
	In    []float64
	Out   []float64
	N     int
	Start time.Time // time of first sample in stream.  This is approximate but normally very close.
}
