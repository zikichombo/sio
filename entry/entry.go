package entry

import (
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

	DefaultSampleCodec sample.Codec
	DefaultBufferSize  int
	DefaultForm        sound.Form
}

// Names returns the (potentially) supported entry point names for the host.
func Names() []string

type entry struct {
	Entry
	pkgPath string
}

var entries []entry

func RegisterEntry(e *Entry) error {
	return nil
}

func EntryFor(name string, pkgSel func(string) *Entry) (*Entry, error) {
	return nil
}

func SourceOpener(pkgSel func(string) bool) {
}

func SinkOpener(pkgSel func(string) bool) {
}

func DuplexOpener(pkgSel func(string) bool) {
}

func DevScanner(pkgSel func(string) bool) {
}

func DevNotifier(pkgSel func(string) bool) {
}
