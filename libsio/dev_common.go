// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package libsio

import (
	"fmt"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

// Dev provides data about an sound device to connect to
// for input or output.
type Dev struct {
	Id             uint64
	Name           string
	SampleCodecs   []sample.Codec
	MaxInChannels  int
	MaxOutChannels int
	MinSampleRate  freq.T
	MaxSampleRate  freq.T
	IsDefaultIn    bool
	IsDefaultOut   bool
	IsDefaultSys   bool
}

func (d *Dev) String() string {
	return fmt.Sprintf("sio.Dev[%d: %s (%d,%d)@[%s..%s] i=%t o=%t s=%t]", d.Id, d.Name,
		d.MaxInChannels, d.MaxOutChannels, d.MinSampleRate, d.MaxSampleRate, d.IsDefaultIn,
		d.IsDefaultOut, d.IsDefaultSys)
}

func (d *Dev) CanInput() bool {
	return d.MaxInChannels > 0
}

func (d *Dev) CanOutput() bool {
	return d.MaxInChannels > 0
}

func (d *Dev) SupportsCodec(c sample.Codec) bool {
	for _, d := range d.SampleCodecs {
		if d == c {
			return true
		}
	}
	return false
}

func (d *Dev) CanOutputForm(v sound.Form) bool {
	nC := v.Channels()
	if nC > d.MaxOutChannels {
		return false
	}
	sr := v.SampleRate()
	if sr < d.MinSampleRate || sr > d.MaxSampleRate {
		return false
	}
	return true
}

func (d *Dev) CanInputForm(v sound.Form) bool {
	nC := v.Channels()
	if nC > d.MaxOutChannels {
		return false
	}
	sr := v.SampleRate()
	if sr < d.MinSampleRate || sr > d.MaxSampleRate {
		return false
	}
	return true
}

func (d *Dev) CanDuplex() bool {
	return d.MaxOutChannels > 0 && d.MaxInChannels > 0
}

func (d *Dev) CanDuplexForm(sr freq.T, inC, outC int) bool {
	if sr < d.MinSampleRate || sr > d.MaxSampleRate {
		return false
	}
	if inC > d.MaxInChannels {
		return false
	}
	if outC > d.MaxOutChannels {
		return false
	}
	return true
}
