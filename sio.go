// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sio

import (
	"zikichombo.org/sio/entry"
	"zikichombo.org/sound"
)

// Capture tries to open the default capture device with
// default settings with the default entry, returning
// a non-nil in case of failure.
func Capture() (sound.Source, error) {
	return nil, nil
}

// Play tries to play a sound.Source
// default settings with the default entry, returning
// a non-nil in case of failure.
func Play(src sound.Source) error {
	ent, err := OpenDefaultEntry(nil)
	if err != nil {
		return err
	}
	return ent.Play(src)
}

// Player tries to return a sound.Sink to which Sends
// are played to some system output.  Default entry
// and settings are applied.
func Player(v sound.Form) (sound.Sink, error) {
	return nil, nil
}

// Duplex tries to return a sound.Duplex.
func Duplex() (sound.Duplex, error) {
	return nil, nil
}

// OpenDefaultEntry tries to return the default entry.
func OpenDefaultEntry(pkgSel func(string) bool) (*entry.Entry, error) {
	return entry.OpenDefault(pkgSel)
}

// OpenEntry tries to return the named entry.
func OpenEntry(name string, pkgSel func(string) bool) (*entry.Entry, error) {
	return entry.Open(name, pkgSel)
}

// Close closes the currently in use entry, if any, so that
// another one may be used.
func Close() {
	entry.Close()
}

// EntryNames returns the names of host entry points.
func EntryNames() []string {
	return entry.Names()
}
