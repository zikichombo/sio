// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build linux
// +build cgo
// +build listen

package sio_test

import (
	"fmt"
	"testing"
	"time"

	"zikichombo.org/sio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/gen"
	"zikichombo.org/sound/sample"
)

func TestAlsaOut(t *testing.T) {
	key := 440 * freq.Hertz
	//z := gen.Notes(key, (key*3)/2, (key*6)/5, 2*key)
	z := gen.Sin(key)
	q, e := sio.DefaultOutputDev.Output(sound.StereoCd(), sample.SFloat32L, 256)
	if e != nil {
		t.Fatal(e)
	}
	defer q.Close()
	fmt.Printf("started q at %s\n", time.Now())
	ttl := 0
	buf := make([]float64, 1)
	start := time.Now()

	for ttl < 44000 {
		pkt, ok := <-q.FillC()
		if !ok {
			break
		}
		pkt.D = pkt.D[:cap(pkt.D)]
		j := 0

		for j < len(pkt.D) {
			n, e := z.Receive(buf)
			if e != nil {
				t.Fatal(e)
			}
			if n != 1 {
				t.Fatalf("expected %d got %d\n", 1, n)
			}
			pkt.D[j], pkt.D[j+1] = buf[0], buf[0]
			j += 2
		}
		ttl += j / 2
		pkt.D = pkt.D[:j]
		q.PlayC() <- pkt
	}
	fmt.Printf("starting delayed at %s.\n", time.Now())
	off := 44000 // half second delay scheduled.
	lim := ttl + 44000
	for ttl < lim {
		pkt, ok := <-q.FillC()
		if !ok {
			break
		}
		pkt.D = pkt.D[:cap(pkt.D)]
		j := 0
		for j < len(pkt.D) {
			n, e := z.Receive(buf)
			if e != nil {
				t.Fatal(e)
			}
			if n != 1 {
				t.Fatalf("expected %d got %d\n", 1, n)
			}
			pkt.D[j], pkt.D[j+1] = buf[0], buf[0]
			j += 2
		}
		if off != 0 {
			pkt.N += off
			off = 0
		}
		ttl += j / 2
		pkt.D = pkt.D[:j]

		q.PlayC() <- pkt
	}
	fmt.Printf("sent %d in %s\n", ttl, time.Since(start))
	time.Sleep(time.Second)
}
