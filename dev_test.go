// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build cgo

package sio_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/irifrance/snd"
	"github.com/irifrance/snd/sio"
)

func TestDevices(t *testing.T) {
	devs := sio.Devices()
	nIns := 0
	nOuts := 0
	for _, d := range devs {
		fmt.Printf("testing %v\n", d)
		for _, sc := range d.SampleCodecs {
			for _, v := range []snd.Form{snd.MonoCd(), snd.StereoCd()} {
				i, ie := d.Input(v, sc, 256)
				if ie != nil {
					t.Errorf("device %s can't make input for %s, %s: %s\n", d, v, sc, ie)
				} else {
					testInput(i, t)
					nIns++
				}
				o, oe := d.Output(v, sc, 256)
				if oe != nil {
					t.Errorf("device %s can't make output for %s, %s: %s\n", d, v, sc, oe)
				} else {
					testOutput(o, t)
					nOuts++
				}
			}
		}
	}
	if nIns == 0 {
		t.Errorf("no inputs found on system.")
	}
	if nOuts == 0 {
		t.Errorf("no outputs found on system.")
	}
}

func testInput(i sio.Input, t *testing.T) {
	i.Close()
	time.Sleep(50 * time.Millisecond)
}

func testOutput(o sio.Output, t *testing.T) {
	o.Close()
	time.Sleep(50 * time.Millisecond)
}
