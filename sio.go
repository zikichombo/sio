package sio

import (
	"zikichombo.org/sio/entry"
	"zikichombo.org/sound"
)

func Capture() (sound.Source, error) {
	return nil, nil
}
func Play() (sound.Source, error) {
	return nil, nil
}
func Player() (sound.Sink, error) {
	return nil, nil
}

func Duplex() (sound.Duplex, error) {
	return nil, nil
}

func OpenDefaultEntry(pkgSel func(string) bool) (*entry.Entry, error) {
	return nil, nil
}

func OpenEntry(name string, pkgSel func(string) bool) (*entry.Entry, error) {
	return nil, nil
}

func Close() {
	entry.Close()
}

func EntryNames() []string {
	return entry.Names()
}
