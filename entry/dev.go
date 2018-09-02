// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package entry

import "zikichombo.org/sio"

// DevScanner scans for devices on the host
type DevScanner interface {
	// DevScan returns a
	DevScan() []*sio.Dev
	// Devices returns a possibly cached list of devices.
	Devices() []*sio.Dev
}

// DevChangeSense indicates whether a DevChange
// is a connection or a disconnection.
type DevChangeSense int

const (
	Connect DevChangeSense = iota
	Disconnect
)

// DevChange describes an event related to
// device connectivity.
type DevChange struct {
	Sense DevChangeSense
	Dev   *sio.Dev
}

// DevNotifier is an interface for notifications of device changes.
type DevNotifier interface {
	// Notify tells the notifier to send notifications of
	// device changes on c.
	//
	// Notify will drop device changes if the channel does not
	// receive the notification, so the caller should buffer c
	// accordingly, for example with a buffer size of length DevScan().
	Notify(c chan<- DevChange)
}
