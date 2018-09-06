package host

import (
	"time"

	"zikichombo.org/sio/libsio"
	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// NullEntry is an entry which implements nothing and
// yet fullfills the Entry interface.
//
// It is useful for embedding in entry points which don't
// implement everything to set defaults.
type NullEntry struct {
}

func (n *NullEntry) Name() string {
	return "null"
}

func (n *NullEntry) DefaultForm() sound.Form {
	return sound.MonoCd()
}

func (n *NullEntry) DefaultSampleCodec() sample.Codec {
	return sample.SFloat32L
}

func (n *NullEntry) DefaultBufSize() int {
	return 256
}

func (n *NullEntry) CanOpenSource() bool {
	return false
}

func (n *NullEntry) OpenSource(d *libsio.Dev, v sound.Form, co sample.Codec, b int) (sound.Source, time.Time, error) {
	var t time.Time
	return nil, t, ErrUnsupported
}

func (n *NullEntry) CanOpenSink() bool {
	return false
}

func (n *NullEntry) OpenSink(d *libsio.Dev, v sound.Form, co sample.Codec, b int) (sound.Sink, *time.Time, error) {
	return nil, nil, ErrUnsupported
}

func (n *NullEntry) CanOpenDuplex() bool {
	return false
}

func (n *NullEntry) OpenDuplex(d *libsio.Dev, iv, ov sound.Form, co sample.Codec, b int) (sound.Duplex, time.Time, *time.Time, error) {
	var t time.Time
	return nil, t, nil, ErrUnsupported
}

func (n *NullEntry) HasDevices() bool {
	return false
}

func (n *NullEntry) ScanDevices() ([]*DevScanResult, error) {
	return nil, ErrUnsupported
}

func (n *NullEntry) Devices() []*libsio.Dev {
	return nil
}

func (n *NullEntry) DevicesNotify(chan<- *DevChange) error {
	return nil
}

func (n *NullEntry) DevicesNotifyClose(c chan<- *DevChange) {
}

func (n *NullEntry) DefaultInputDev() *libsio.Dev {
	return nil
}

func (n *NullEntry) DefaultOutputDev() *libsio.Dev {
	return nil
}

func (n *NullEntry) DefaultDuplexDev() *libsio.Dev {
	return nil
}
