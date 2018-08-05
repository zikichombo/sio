// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build darwin

package sio_test

import (
	"testing"
	"time"

	"github.com/irifrance/snd"
	"github.com/irifrance/snd/encoding/wav"
	"github.com/irifrance/snd/ops"
	"github.com/irifrance/snd/sample"
	"github.com/irifrance/snd/sio"
)

func TestDarwinIn(t *testing.T) {
	v := snd.StereoCd()
	q, e := sio.DefaultInputDev.Input(v, sample.SFloat32L, 128)
	if e != nil {
		t.Fatal(e)
	}
	defer q.Close()
	src := ops.LimitDur(sio.InputSource(q), time.Second)
	if e := wav.Save(src, "darwin-in.wav"); e != nil {
		t.Fatal(e)
	}
}
