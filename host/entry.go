// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package host

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/ops"
	"zikichombo.org/sound/sample"
)

// Entry is the type of a connection to host sound I/O capabilities.
//
// Entries may either be high level (eg AudioFlinger) or low level (eg ALSA hw).
//
// Multiple entries may exist for a given host.  see Names() in the relevant
// entry_{runtime.GOOS}.go file for details.
type Entry interface {
	// Name returns the name of the entry and should be a valid
	// name for the host.
	Name() string
	// DefaultForm returns the default form.  In case one form is not
	// implemented or uniform accross all supported functionality (input, output, device),
	// output default should be returned.
	DefaultForm() sound.Form

	// DefaultSampleCodec returns the default sample codec.  In case one sample codec
	// is not implemented or uniform accross all supported functionality, output default
	// should be returned.
	DefaultSampleCodec() sample.Codec

	// DefaultBufSize returns the default size of the buffer in frames provided to
	// the caller for I/O interactions.
	DefaultBufSize() int

	// CanOpenSource returns true if OpenSource does not return
	// ErrUnspported.
	CanOpenSource() bool

	// OpenSource starts capturing audio.
	//
	// dev is the specified device, and should be nil if and only if the current
	// entry has .HasDevices().  Otherwise, dev should be a device returned by
	// the current entry .ScanDevices() or Devices()
	//
	// v indicates the desired form (channels, sample rate) of the source.
	//
	// sco indicates the desired sample.Codec.
	//
	// bufSz indicates the size of buffer, in frames, which data from the
	// hardware or software implementing the entry to sound.Source.Receive.
	// For hardware, this is normally the size of the part of the ring buffer exposed for
	// reading.  Implementations should use a minimal total buffer size to accomodate
	// this constraint.
	//
	// OpenSource returns a triple (s, t, e) with
	// s: sound.Source which represents captured audio.
	// t: the start time of the first sample.
	// e: any error
	OpenSource(dev *libsio.Dev, v sound.Form, sco sample.Codec, bufSz int) (sound.Source, time.Time, error)

	// Can open sink returns true if OpenSink does not return
	// ErrUnsupported
	CanOpenSink() bool
	// OpenSink starts playing audio.
	//
	// dev is the specified device, and should be nil if and only if the current
	// entry has .HasDevices().  Otherwise, dev should be a device returned by
	// the current entry via ScanDevices() or Devices().
	//
	// v indicates the desired form (channels, sample rate) of the source.
	//
	// sco indicates the desired sample.Codec.
	//
	// bufSz indicates the size of buffer whose data is placed
	// in sound.Source.Receive.  This is normally the size of the part of
	// the ring buffer exposed for playback.  Implementations should
	// use a minimal total buffer size to safely accomodate this constraint.
	//
	// OpenSink returns a tuple (s, *t, e) with
	// s: sound.Sink which represents audio for playback.
	// *t: a pointer to the start time of the first played sample, which is set
	//    after the first successful send via the returned sound.Sink.
	// e: any error
	//
	// OpenSink should
	OpenSink(dev *libsio.Dev, v sound.Form, sco sample.Codec, bufSz int) (sound.Sink, *time.Time, error)

	// CanOpenDuplex returns true if OpenDuplex does not return ErrUnsupported.
	CanOpenDuplex() bool
	// OpenDuplex starts a duplex connection to the host.
	//
	// dev is the specified device, and should be nil if and only if the current
	// entry has .HasDevices().  Otherwise, dev should be a device returned by
	// the current entry via ScanDevices() or Devices().
	//
	// iv, ov represent the form of input and output respectively.
	//
	// sco indicates the desired sample.Codec.
	//
	// bufSz indicates the size of the buffer, in frames, whose data is
	// placed in or retreived from a []float64 in Duplex.SendReceive. Implementations
	// should use a minimal total buffer size to safely accomodate this constraint.
	//
	// OpenDuplex returns a quadruple(d, ct, *pt, e) with
	//
	//  d a sound.Duplex implementation
	//  ct the time of the first sample of captured data.
	//  *pt: a pointer to the time of the first sample of played data.
	//  e: an error if any while opening the duplex connection.
	OpenDuplex(dev *libsio.Dev, iv, ov sound.Form, sco sample.Codec, bufSz int) (sound.Duplex, time.Time, *time.Time, error)

	// HasDevices returns true if the entry supports the concept of devices.
	HasDevices() bool

	// ScanDevices scans the host for devices.  It does not use a cache.
	//
	// ScanDevices returns a non-nil error if no devices can be scanned.
	// Errors associated with each device scan are available in the
	// DevScanResults.
	ScanDevices() ([]*DevScanResult, error)

	// Devices returns a slice of devices on the host.  The returned list may be cached.
	// Devices() should call ScanDevices() at least once and cache the resulting slice
	// of devices subsequently.
	Devices() []*libsio.Dev

	// DevicesNotify  sends notifications of device changes on c.
	//
	// DevicesNotify returns ErrUnsupported if the entry does not support
	// notifications.
	DevicesNotify(c chan<- *DevChange) error

	// DevicesNotifyClose stops sending device notifications on c.
	DevicesNotifyClose(c chan<- *DevChange)

	// returns nil if no input supported or HasDevices() is false.
	DefaultInputDev() *libsio.Dev
	// returns nil if no output supported or HasDevices() is false.
	DefaultOutputDev() *libsio.Dev
	// returns nil if no duplex supported or HasDevices() is false.
	DefaultDuplexDev() *libsio.Dev
}

// DevScanResult
type DevScanResult struct {
	Dev *libsio.Dev
	E   error
}

// DevChangeSense indicates whether a DevChange
// is a connection or a disconnection.
type DevChangeSense int

const (
	DeviceConnect DevChangeSense = iota
	DeviceDisconnect
)

// DevChange describes an event related to
// device connectivity.
type DevChange struct {
	Sense DevChangeSense
	Dev   *libsio.Dev
}

// Capture tries to open an input such as a microphone.
func Capture(e Entry) (sound.Source, error) {
	return CaptureWith(e, e.DefaultForm(), e.DefaultSampleCodec(), e.DefaultBufSize())
}

// CaptureWith opens audio capture such as via a microphone
// with the specified sound.Form v, sample.Codec co, and buffer size b.
func CaptureWith(e Entry, v sound.Form, co sample.Codec, b int) (sound.Source, error) {
	if !e.CanOpenSource() {
		return nil, ErrUnsupported
	}
	dev := e.DefaultInputDev()
	s, _, err := e.OpenSource(dev, v, co, b)
	return s, err
}

// Play plays a sound.Source
func Play(e Entry, src sound.Source) error {
	snk, err := Player(e, src)
	if err != nil {
		return err
	}
	return ops.Copy(snk, src)
}

// PlayWith returns a Player for src with output sample codec co
// and buffer size b.
func PlayWith(e Entry, src sound.Source, co sample.Codec, b int) error {
	snk, err := PlayerWith(e, src, co, b)
	if err != nil {
		return err
	}
	return ops.Copy(snk, src)
}

// Player tries to return a sound.Sink to which
// writes are played.
func Player(e Entry, v sound.Form) (sound.Sink, error) {
	return PlayerWith(e, v, e.DefaultSampleCodec(), e.DefaultBufSize())
}

// PlayerWith opens a sound.Sink for playback with the specified sound.Form,
// sample.Codec, and buffersize.
func PlayerWith(e Entry, v sound.Form, co sample.Codec, b int) (sound.Sink, error) {
	if !e.CanOpenSink() {
		return nil, ErrUnsupported
	}
	dev := e.DefaultOutputDev()
	snk, _, err := e.OpenSink(dev, v, co, b)
	if err != nil {
		return nil, err
	}
	return snk, nil
}

// Duplex returns a duplex with entry defaults.
func Duplex(e Entry, iv, ov sound.Form) (sound.Duplex, error) {
	return DuplexWith(e, iv, ov, e.DefaultSampleCodec(), e.DefaultBufSize())
}

// DuplexWith opens a Duplex connection with the specified input, output forms,
// sample codec, and buffer size.
func DuplexWith(e Entry, iv, ov sound.Form, co sample.Codec, b int) (sound.Duplex, error) {
	if !e.CanOpenDuplex() {
		return nil, ErrUnsupported
	}
	dev := e.DefaultDuplexDev()
	dpx, _, _, err := e.OpenDuplex(dev, iv, ov, co, b)
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

var entries map[string][]*entry = make(map[string][]*entry)

// RegisterEntry registers an Entry.
func RegisterEntry(e Entry) error {
	nm := e.Name()
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
		Entry:   e,
		pkgPath: pkgPath(e)}
	entries[nm] = append(entries[nm], privEntry)
	return nil
}

var hMu sync.Mutex
var eMu sync.Mutex
var theEntry Entry

// Connect opens the default entry for the host.
//
// Connect returns ErrNoEntryAvailable if there are no entries for the host.
//
// Connect returns ErrEntryInUse if a non-default host entry is in use.
//
// Connect can be called many times without cost of re-initialising a
// connection.  Connect can be called in different goroutines.
func Connect(pkgSel func(string) bool) (Entry, error) {
	nms := Names()
	if len(nms) == 0 {
		return nil, ErrNoEntryAvailable
	}
	return ConnectTo(nms[0], pkgSel)
}

// ConnectTo connects to the named entry.
//
// ConnectTo returns ErrNoEntryAvailable if there are no entries for the host.
//
// ConnectTo returns ErrEntryInUse if another host entry other than one
// requested is in use.
//
// Connect can be called many times without cost of re-initialising a
// connection.  Connect can be called in different goroutines.
func ConnectTo(name string, pkgSel func(string) bool) (Entry, error) {
	hMu.Lock()
	defer hMu.Unlock()
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

// Disconnect disconnects any current entry making
// other entries available.
//
// Disconnect is safe for calling in several goroutines.
func Disconnect() {
	hMu.Lock()
	defer hMu.Unlock()
	if theEntry != nil {
		theEntry = nil
		eMu.Unlock()
	}
}

func findEntry(s []*entry, pkgSel func(string) bool) Entry {
	for _, e := range s {
		if pkgSel == nil || pkgSel(e.Name()) {
			return e.Entry
		}
	}
	return nil
}

// Names names the sound system entry points for the host.
func Names() []string {
	res := make([]string, len(names))
	copy(res, names[:])
	return res
}
