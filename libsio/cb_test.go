// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package libsio

import (
	"fmt"
	"testing"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

func TestCbCapture(t *testing.T) {
	N := 1024
	v := sound.MonoCd()
	c := sample.SFloat32L
	b := 512
	cb := NewCb(v, c, b)
	go runcbsCapture(cb, N, b, c.Bytes())
	d := make([]float64, b)
	for i := 0; i < N; i++ {
		fmt.Printf("cb receive %d\n", i)
		n, err := cb.Receive(d)
		if err != nil {
			t.Error(err)
		} else if n != b {
			t.Errorf("expected %d got %d\n", b, n)
		}
	}
}

func TestCbPlay(t *testing.T) {
	N := 1024
	v := sound.MonoCd()
	c := sample.SFloat32L
	b := 512
	cb := NewCb(v, c, b)
	go runcbsPlay(cb, N, b, c.Bytes())
	d := make([]float64, b)
	for i := 0; i < N; i++ {
		fmt.Printf("cb send %d\n", i)
		err := cb.Send(d)
		if err != nil {
			t.Error(err)
		}
	}
}
