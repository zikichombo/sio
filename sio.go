// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

import (
	"zikichombo.org/sio/host"
	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// Capture tries to open the default capture device with
// default settings with the default host, returning
// a non-nil in case of failure.
func Capture() (sound.Source, error) {
	ent, err := Connect(nil)
	if err != nil {
		return nil, err
	}
	return host.Capture(ent)
}

// CaptureWith opens a sound.Source with the specified sample codec and
// buffer size.
func CaptureWith(v sound.Form, co sample.Codec, b int) (sound.Source, error) {
	ent, err := Connect(nil)
	if err != nil {
		return nil, err
	}
	return host.CaptureWith(ent, v, co, b)
}

// Play tries to play a sound.Source
// default settings with the default entry, returning
// a non-nil in case of failure.
func Play(src sound.Source) error {
	ent, err := Connect(nil)
	if err != nil {
		return err
	}
	return host.Play(ent, src)
}

// PlayWith
func PlayWith(src sound.Source, co sample.Codec, b int) error {
	ent, err := Connect(nil)
	if err != nil {
		return err
	}
	return host.PlayWith(ent, src, co, b)
}

// Player tries to return a sound.Sink to which Sends
// are played to some system output.  Default entry
// and settings are applied.
func Player(v sound.Form) (sound.Sink, error) {
	ent, err := Connect(nil)
	if err != nil {
		return nil, err
	}
	return host.Player(ent, v)
}

// PlayerWith tries to return a sound.Sink for playback
// with the specified sample codec and buffer size b.
func PlayerWith(v sound.Form, co sample.Codec, b int) (sound.Sink, error) {
	ent, err := Connect(nil)
	if err != nil {
		return nil, err
	}
	return host.PlayerWith(ent, v, co, b)
}

// Duplex tries to return a sound.Duplex.
func Duplex(in, out sound.Form) (sound.Duplex, error) {
	ent, err := Connect(nil)
	if err != nil {
		return nil, err
	}
	return host.Duplex(ent, in, out)
}

// DuplexWith tries to return a sound.Duplex.
func DuplexWith(in, out sound.Form, co sample.Codec, b int) (sound.Duplex, error) {
	ent, err := Connect(nil)
	if err != nil {
		return nil, err
	}
	return host.DuplexWith(ent, in, out, co, b)
}

// Connect returns a connection to the default host sound system entry
// point "entry".
//
// Connect returns ErrNoEntryAvailable if there are no entries for the host.
//
// Connect returns ErrEntryInUse if a non-default host entry is in use.
//
// Connect can be called many times without cost of re-initialising a
// connection.  Connect can be called in different goroutines.
//
// The argument pkgSel is used to filter the implementions of
// an entry point by means of examining their defining package path
// by reflection.  It should return true if it accepts the
// implementation.  If pkgSel is nil, Connect acts as though
// the function body were "return true".
func Connect(pkgSel func(string) bool) (host.Entry, error) {
	return host.Connect(pkgSel)
}

// ConnectTo connects to the named host sound system entry point "entry".
//
// ConnectTo returns ErrNoEntryAvailable if there are no entries for the host.
//
// ConnectTo returns ErrEntryInUse if another host entry other than one
// requested is in use.
//
// ConnectTo can be called many times without cost of re-initialising a
// connection.  ConnectTo can be called in different goroutines.
//
// pkgSel is as in Connect.
func ConnectTo(name string, pkgSel func(string) bool) (host.Entry, error) {
	return host.ConnectTo(name, pkgSel)
}

// Disconnect closes the currently in use entry, if any, so that
// another one may be used.
func Disconnect() {
	host.Disconnect()
}

// EntryNames returns the names of host entry points.
func EntryNames() []string {
	return host.Names()
}
