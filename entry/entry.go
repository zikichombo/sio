package entry

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"zikichombo.org/sio"
	"zikichombo.org/sound"
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
	// {Source,Sink,Duplex}Opener for the entry.
	DefaultBufferSize int
	// Default Form gives a default sound form, as passed to
	// {Source,Sink,Duplex}Opener for the entry.
	DefaultForm sound.Form
}

// Capture tries to open an input such as a microphone.
func (e *Entry) Capture() (sound.Source, error) {
	if e.SourceOpener == nil {
		return nil, ErrUnsupported
	}
	var theDev *sio.Dev
	if e.DevScan != nil && len(e.Devices()) > 0 {
		theDev = e.Devices()[0]
	}
	s, _, err := e.SourceOpener(theDev, e.DefaultForm, e.DefaultSampleCodec, e.DefaultBufferSize)
	return s, err
}

// Play plays a sound.Source
func (e *Entry) Play(s sound.Source) error {
	if e.SinkOpener == nil {
		return nil, ErrUnsupported
	}
	var theDev *sio.Dev
	if e.DevScan != nil && len(e.Devices()) > 0 {
		theDev = e.Devices()[0]
	}
	snk, _, err := e.SinkOpener(theDev, e.DefaultForm, e.DefaultSampleCodec, e.DefaultBufferSize)
	return snk, err
}

// Duplex returns a duplex with entry defaults.
func (e *Entry) Duplex() (sound.Duplex, error) {
	if e.SinkOpener == nil {
		return nil, ErrUnsupported
	}
	var theDev *sio.Dev
	if e.DevScan != nil && len(e.Devices()) > 0 {
		theDev = e.Devices()[0]
	}
	dpx, _, err := e.DuplexOpener(theDev, e.DefaultForm, e.DefaultSampleCodec, e.DefaultBufferSize)
	return dpx, err
}

// Names returns the (potentially) supported entry point names for the host.
// The names are taken in priority order.
func Names() []string

type entry struct {
	Entry
	pkgPath string
}

func pkgPath(v interface{}) string {
	typ := reflect.ValueOf(v).Type()
	return typ.PkgPath()
}

var entries map[string][]*entry

// ErrInvalidEntryName is used on RegisterEntry to
// enforce the use of known entry points.
var ErrInvalidEntryName = errors.New("invalid entry name")

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

func OpenDefault(pkgSel func(string) *Entry) (*Entry, error) {
	nms := Names()
	if len(nms) == 0 {
		return nil, ErrNoEntryAvailable
	}
	return Open(nms[0])
}

// Open
func Open(name string, pkgSel func(string) *Entry) (*Entry, error) {
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
