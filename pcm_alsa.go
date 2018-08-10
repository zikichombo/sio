// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// +build linux
// +build cgo

package sio

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	"unsafe"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

// #cgo LDFLAGS: -lasound
// #include "alsa/asoundlib.h"
//
import "C"

type alsaPcm struct {
	sound.Form
	codec      sample.Codec
	name       string
	pcm        *C.snd_pcm_t
	dir        C.snd_pcm_stream_t
	hwParams   *C.snd_pcm_hw_params_t
	swParams   *C.snd_pcm_sw_params_t
	perBuf     *C.char
	start      time.Time
	pkts       [3]Packet
	pi         int
	doneC      chan struct{}
	pktC       [2]chan *Packet
	once       sync.Once
	periodSize C.ulong
	periods    int
}

func newAlsaPcmIn(name string, v sound.Form, sc sample.Codec, nf int) *alsaPcm {
	res := newAlsaPcm(name, v, sc, nf)
	res.dir = C.SND_PCM_STREAM_CAPTURE
	res.pktC[0] = make(chan *Packet, 1)
	return res
}

func newAlsaPcmOut(name string, v sound.Form, sc sample.Codec, nf int) *alsaPcm {
	res := newAlsaPcm(name, v, sc, nf)
	res.dir = C.SND_PCM_STREAM_PLAYBACK
	res.pktC[0] = make(chan *Packet, 1)
	res.pktC[1] = make(chan *Packet)
	return res
}

func newAlsaPcm(name string, v sound.Form, sc sample.Codec, nf int) *alsaPcm {
	res := &alsaPcm{
		Form:  v,
		codec: sc,
		name:  name}
	res.doneC = make(chan struct{})
	res.periodSize = C.ulong(nf)
	return res
}

func (dev *alsaPcm) open() error {
	cName := C.CString(dev.name)
	ret := C.snd_pcm_open(&dev.pcm, cName, dev.dir, 0)
	defer C.free(unsafe.Pointer(cName))
	if ret < 0 {
		return fmt.Errorf("unable to open device: %s", sndStrerror(ret).Error())
	}

	C.snd_pcm_hw_params_malloc(&dev.hwParams)
	C.snd_pcm_hw_params_any(dev.pcm, dev.hwParams)
	ret = C.snd_pcm_hw_params_set_access(dev.pcm, dev.hwParams,
		C.SND_PCM_ACCESS_RW_INTERLEAVED)
	if ret < 0 {
		return fmt.Errorf("unable to set access: %s",
			sndStrerror(ret))
	}
	ret = C.snd_pcm_hw_params_set_channels(dev.pcm, dev.hwParams, C.uint(dev.Channels()))
	if ret < 0 {
		return fmt.Errorf("unable to set number of channels to %d: %s",
			dev.Channels(), sndStrerror(ret))
	}
	C.snd_pcm_hw_params_set_format(dev.pcm, dev.hwParams, scodec2Alsa[dev.codec])
	if ret < 0 {
		return fmt.Errorf("unable to set sample codec to %s: %s",
			dev.codec, sndStrerror(ret))
	}
	var unusedDir C.int
	ret = C.snd_pcm_hw_params_set_rate(dev.pcm, dev.hwParams,
		C.uint(dev.SampleRate()/freq.Hertz), unusedDir)
	if ret < 0 {
		return fmt.Errorf("unable to set sample rate to to %s: %s",
			dev.SampleRate(), sndStrerror(ret))
	}
	if err := dev.chooseBufPeriods(int(dev.periodSize)); err != nil {
		return err
	}
	ret = C.snd_pcm_hw_params(dev.pcm, dev.hwParams)
	if ret < 0 {
		return fmt.Errorf("unable to set hw params: %s", sndStrerror(ret))
	}

	if dev.dir == C.SND_PCM_STREAM_CAPTURE {
		go dev.serveCapture()
	} else {
		go dev.servePlay()
	}
	return nil
}

func (dev *alsaPcm) bufFrames() int {
	return len(dev.pkts[0].D) / dev.Channels()
}

func (dev *alsaPcm) chooseBufPeriods(nF int) error {
	pMin := C.ulong(0)
	pMax := C.ulong(0)
	minDir := C.int(0)
	maxDir := C.int(0)

	C.snd_pcm_hw_params_get_period_size_min(dev.hwParams, &pMin, &minDir)
	C.snd_pcm_hw_params_get_period_size_max(dev.hwParams, &pMax, &maxDir)
	if minDir > 0 {
		pMin++
	}
	if maxDir < 0 {
		pMax--
	}
	per := C.ulong(nF)
	if per < pMin {
		log.Printf("alsa buffer (period) size %d unavailable, using %d\n", per, pMin)
		per = pMin
	}
	if per > pMax {
		log.Printf("alsa buffer (period) size %d unavailable, using %d\n", per, pMax)
		per = pMax
	}
	var ret C.int
	ret = C.snd_pcm_hw_params_set_period_size(dev.pcm, dev.hwParams, per, 0)
	if ret < 0 {
		return fmt.Errorf("unable to set period size to %d: %s",
			per, sndStrerror(ret))
	}
	dev.periodSize = per
	ns := int(per) * dev.Channels()
	for i := 0; i < 3; i++ {
		dev.pkts[i].D = make([]float64, ns)
	}

	bufMin := C.snd_pcm_uframes_t(0)
	bufMax := C.snd_pcm_uframes_t(0)

	C.snd_pcm_hw_params_get_buffer_size_min(dev.hwParams, &bufMin)
	C.snd_pcm_hw_params_get_buffer_size_max(dev.hwParams, &bufMax)
	buf := 3 * dev.periodSize
	if buf > bufMax {
		return fmt.Errorf("buffer size would be forced to %d, need %d",
			bufMax, buf)
	}
	if buf < bufMin {
		log.Printf("buffer size forced to %d", bufMin)
		buf = bufMin
	}
	ret = C.snd_pcm_hw_params_set_buffer_size(dev.pcm, dev.hwParams, buf)
	if ret < 0 {
		return fmt.Errorf("unable to set buffer size to %d: %s",
			buf, sndStrerror(ret))
	}
	pszBytes := C.ulong(dev.Channels()*dev.codec.Bytes()) * per
	dev.perBuf = (*C.char)(C.malloc(pszBytes))
	dev.periods = int(buf) / int(per)
	return nil
}

func (dev *alsaPcm) serveCapture() {
	N := 0
	codec := dev.codec
	bytesPerFrame := C.long(dev.Channels() * codec.Bytes())
	pi := 0
	start := time.Now()
	for i := range dev.pkts {
		dev.pkts[i].Start = start
	}
	defer dev.pcmClose()
	buf := unsafe.Pointer(dev.perBuf)

	for {
		nf := C.snd_pcm_readi(dev.pcm, buf, dev.periodSize)
		switch nf {
		case -C.EPIPE:
			log.Printf("alsa: overrun")
			C.snd_pcm_prepare(dev.pcm)
			continue
		case -C.EBADFD:
			log.Printf("alsa: bad state")
			C.snd_pcm_prepare(dev.pcm)
			continue
		case -C.ESTRPIPE:
			log.Printf("estrpipe: driver suspended")
			return
		case 0:
			return
		}
		pkt := &dev.pkts[pi]
		slice := (*[1 << 30]byte)(buf)[:nf*bytesPerFrame]
		pkt.D = pkt.D[:int(nf)*dev.Channels()]
		codec.Decode(pkt.D, slice)
		pkt.N = N
		N += int(nf)
		select {
		case <-dev.doneC:
			return
		case dev.pktC[0] <- pkt:
		}
		pi++
		if pi == len(dev.pkts) {
			pi = 0
		}
	}
}

func (dev *alsaPcm) servePlay() {
	defer dev.pcmClose()
	buf := unsafe.Pointer(dev.perBuf)
	nC := dev.Channels()
	bytesPerFrame := dev.codec.Bytes() * dev.Channels()
	codec := dev.codec
	start := time.Now()
	for i := range dev.pkts {
		dev.pkts[i].Start = start
	}
	dev.initPerBuf()
	// prime the device loop
	for i := 0; i < dev.periods; i++ {
		if err := dev.writei(dev.periodSize); err != nil {
			log.Printf("error: %s\n", err)
			return
		}
	}
	var pkt *Packet
	var ok bool
	pi := 0
	N := 0
	for {
		pkt = &dev.pkts[pi]
		pkt.N = N
		select {
		case <-dev.doneC:
			return
		case dev.pktC[0] <- pkt:
		}
		select {
		case <-dev.doneC:
			return
		case pkt, ok = <-dev.pktC[1]:
			if !ok {
				// closed PlayC() -> Close()
				return
			}
		}
		// check memory reqs respected
		if &pkt.D[0] != &dev.pkts[pi].D[0] {
			panic("must use packet memory")
		}
		pi++
		if pi == len(dev.pkts) {
			pi = 0
		}
		// check non monotonic frame number
		if pkt.N < N {
			log.Printf("alsa: non monotonic frame number schedule")
			pkt.N = N
		}
		// check scheduling in the future.
		if pkt.N > N {
			if err := dev.writeSilence(C.ulong(pkt.N - N)); err != nil {
				log.Printf("alsa: error: %s\n", err)
				return
			}
			N = pkt.N
		}
		nF := len(pkt.D) / nC
		slice := (*[1 << 30]byte)(buf)[:nF*bytesPerFrame]
		codec.Encode(slice, pkt.D)
		if err := dev.writei(dev.periodSize); err != nil {
			log.Printf("error: %s\n", err)
			return
		}
		N += int(dev.periodSize)
	}
}

func (dev *alsaPcm) initPerBuf() {
	M := int(dev.periodSize) * dev.Channels() * dev.codec.Bytes()
	slice := (*[1 << 30]byte)(unsafe.Pointer(dev.perBuf))[:M]
	for i := 0; i < M; i++ {
		slice[i] = 0
	}
}

func (dev *alsaPcm) writeSilence(n C.ulong) error {
	dev.initPerBuf()
	var i C.ulong
	for i < n {
		m := n - i
		if m > dev.periodSize {
			m = dev.periodSize
		}
		if err := dev.writei(m); err != nil {
			return err
		}
		i += m
	}
	return nil
}

func (dev *alsaPcm) writei(wf C.ulong) error {
wi:
	nf := C.snd_pcm_writei(dev.pcm, unsafe.Pointer(dev.perBuf), wf)
	switch nf {
	case -C.EPIPE:
		log.Printf("alsa: underrun")
		C.snd_pcm_prepare(dev.pcm)
		goto wi
	case -C.EBADFD:
		log.Printf("alsa: bad pcm state")
		C.snd_pcm_prepare(dev.pcm)
		goto wi
	case -C.ESTRPIPE:
		log.Printf("alsa: driver suspended, exiting")
		return fmt.Errorf("driver suspended")
	}
	return nil
}

func (dev *alsaPcm) C() <-chan *Packet {
	return dev.pktC[0]
}

func (dev *alsaPcm) FillC() <-chan *Packet {
	return dev.pktC[0]
}

func (dev *alsaPcm) PlayC() chan<- *Packet {
	return dev.pktC[1]
}

func (dev *alsaPcm) Close() error {
	dev.once.Do(func() { close(dev.doneC) })
	return nil
}

func (dev *alsaPcm) pcmClose() {
	if dev.perBuf != nil {
		C.free(unsafe.Pointer(dev.perBuf))
	}
	C.snd_pcm_drain(dev.pcm)
	C.snd_pcm_close(dev.pcm)
	C.snd_pcm_hw_params_free(dev.hwParams)
	C.snd_pcm_sw_params_free(dev.swParams)
}

func sndStrerror(c C.int) error {
	return errors.New(C.GoString(C.snd_strerror(c)))
}
