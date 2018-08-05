// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build linux
// +build cgo

package sio

import (
	"testing"
	"time"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

func TestAlsaOpen(t *testing.T) {
	for _, c := range []sample.Codec{sample.SInt16L} { //sample.Codecs {
		for _, v := range []sound.Form{sound.StereoCd(), sound.MonoCd()} {
			in := newAlsaPcmIn("default", v, c, 256)
			if err := in.open(); err != nil {
				t.Error(err)
				in.Close()
				continue
			}
			time.Sleep(time.Second)
			in.Close()
			out := newAlsaPcmOut("default", v, c, 128)
			if err := out.open(); err != nil {
				t.Error(err)
				out.Close()
				continue
			}
			out.Close()
		}
	}
}
