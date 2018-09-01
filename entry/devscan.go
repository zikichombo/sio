package entry

import "zikichombo.org/sio"

type DevScanner interface {
	DevScan() []*sio.Dev
}
