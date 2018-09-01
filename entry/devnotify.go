package entry

import "zikichombo.org/sio"

type DevChangeSense int

const (
	Connect DevChangeSense = iota
	Disconnect
)

type DevChange struct {
	Sense DevChangeSense
	Dev   *sio.Dev
}

type DevNotifier interface {
	Notify(chan<- DevChange)
}
