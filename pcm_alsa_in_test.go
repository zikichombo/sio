// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build cgo
// +build linux

package sio_test

import (
	"testing"
	"time"

	"zikichombo.org/sound"
	"zikichombo.org/sound/encoding/wav"
	"zikichombo.org/sound/ops"
	"zikichombo.org/sound/sample"
	"zikichombo.org/sound/sio"
)

func TestAlsaIn(t *testing.T) {
	v := sound.StereoCd()
	q, e := sio.DefaultInputDev.Input(v, sample.SFloat32L, 128)
	if e != nil {
		t.Fatal(e)
	}
	defer q.Close()
	src := ops.LimitDur(sio.InputSource(q), 4*time.Second)
	if e := wav.Save(src, "alsa-in.wav"); e != nil {
		t.Fatal(e)
	}
}
