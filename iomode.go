// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

// IoMode represents direction of audio data in a device
type IoMode int

const (
	InputMode  IoMode = iota // audio capture
	OutputMode               // audio playback
	DuplexMode               // both capture and playback synchronised.
)

// Inputs returns true iff d is either InputMode or DuplexMode
func (d IoMode) Inputs() bool {
	return d == InputMode || d == DuplexMode
}

// Outputs returns true iff d is either OutputMode or DuplexMode
func (d IoMode) Outputs() bool {
	return d == OutputMode || d == InputMode
}

// Duplex returns true iff d is DuplexMode
func (d IoMode) Duplex() bool {
	return d == DuplexMode
}
