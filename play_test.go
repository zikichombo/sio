// +build listen

package sio_test

import (
	"testing"
	"time"

	"zikichombo.org/sio"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/gen"
	"zikichombo.org/sound/ops"
)

func TestPlay(t *testing.T) {
	src := ops.LimitDur(gen.Note(440*freq.Hertz), time.Second)
	if err := sio.Play(src); err != nil {
		t.Fatal(err)
	}
}
