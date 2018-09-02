// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package entry

import (
	"fmt"
	"reflect"
	"sync"

	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/ops"
	"zikichombo.org/sound/sample"
)

// An Entry is an way of getting into the sound system of
// the host.
type Entry struct {
	// Name should be one of the supported names (Names())
	Name string
	// SourceOpener should provide capture capabilities or be nil.
	SourceOpener
	// SinkOpener should provide playback capabilities or be nil.
	SinkOpener
	// DuplexOpener should provide duplex capabilities or be nil
	DuplexOpener

	// DevScanner should provide device scanning capabilities or be nil.
	DevScanner
	// DevNotifier should provide device notification capabilities or be nil.
	DevNotifier DevNotifier

	// DefaultSampleCodec gives a default sample codec for whatever
	// functionality is supported by the Entry.
	DefaultSampleCodec sample.Codec

	// DefaultBufferSize gives a default buffer size, as passed to
	// {Source}Opener for the entry.
	DefaultInputBufferSize  int
	DefaultOutputBufferSize int
	DefaultDuplexBufferSize int
	// Default Form gives a default sound form, as passed to
	// {Source,Sink,Duplex}Opener for the entry.
	DefaultForm sound.Form
}

// Capture tries to open an input such as a microphone.
func (e *Entry) Capture() (sound.Source, error) {
	if e.SourceOpener == nil {
		return nil, ErrUnsupported
	}
	var theDev *libsio.Dev
	if e.DevScan != nil && len(e.Devices()) > 0 {
		theDev = e.Devices()[0]
	}
	s, _, err := e.SourceOpener.OpenSource(theDev, e.DefaultForm,
		e.DefaultSampleCodec, e.DefaultInputBufferSize)
	return s, err
}

// Play plays a sound.Source
func (e *Entry) Play(src sound.Source) error {
	snk, err := e.Player()
	if err != nil {
		return err
	}
	return ops.Copy(snk, src)
}

// Player tries to return a sound.Sink to which
// writes are played.
func (e *Entry) Player() (sound.Sink, error) {
	if e.SinkOpener == nil {
		return nil, ErrUnsupported
	}
	var theDev *libsio.Dev
	if e.DevScan != nil && len(e.Devices()) > 0 {
		theDev = e.Devices()[0]
	}
	snk, _, err := e.SinkOpener.OpenSink(theDev, e.DefaultForm,
		e.DefaultSampleCodec, e.DefaultOutputBufferSize)
	if err != nil {
		return nil, err
	}
	return snk, err
}

// Duplex returns a duplex with entry defaults.
func (e *Entry) Duplex() (sound.Duplex, error) {
	if e.SinkOpener == nil {
		return nil, ErrUnsupported
	}
	var theDev *libsio.Dev
	if e.DevScan != nil && len(e.Devices()) > 0 {
		theDev = e.Devices()[0]
	}
	dpx, _, err := e.DuplexOpener.OpenDuplex(theDev, e.DefaultForm,
		e.DefaultSampleCodec, e.DefaultDuplexBufferSize)
	return dpx, err
}

type entry struct {
	Entry
	pkgPath string
}

func pkgPath(v interface{}) string {
	typ := reflect.ValueOf(v).Type()
	return typ.PkgPath()
}

var entries map[string][]*entry

// RegisterEntry registers an Entry
func RegisterEntry(e *Entry) error {
	nm := e.Name
	found := false
	for _, okName := range Names() {
		if nm == okName {
			found = true
			break
		}
	}
	if !found {
		return ErrInvalidEntryName
	}
	privEntry := &entry{
		Entry:   *e,
		pkgPath: pkgPath(e)}
	entries[nm] = append(entries[nm], privEntry)
	return nil
}

var eMu sync.Mutex
var theEntry *Entry

func OpenDefault(pkgSel func(string) bool) (*Entry, error) {
	nms := Names()
	if len(nms) == 0 {
		return nil, ErrNoEntryAvailable
	}
	return Open(nms[0], pkgSel)
}

// Open
func Open(name string, pkgSel func(string) bool) (*Entry, error) {
	if theEntry != nil {
		return theEntry, nil
	}
	res := findEntry(entries[name], pkgSel)
	if res == nil {
		return nil, fmt.Errorf("couldn't locate entry.")
	}
	if theEntry != nil && theEntry != res {
		return nil, ErrEntryInUse
	}
	eMu.Lock()
	theEntry = res
	return res, nil
}

// TBD(wsc) deal with open/close race.
func Close() {
	if theEntry != nil {
		theEntry = nil
		eMu.Unlock()
	}
}

func findEntry(s []*entry, pkgSel func(string) bool) *Entry {
	for _, e := range s {
		if pkgSel == nil || pkgSel(e.Name) {
			return &e.Entry
		}
	}
	return nil
}
