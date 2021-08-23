//go:build listen
// +build listen

package sio_test

import (
	"testing"
	"time"

	"github.com/zikichombo/sio"
	"github.com/zikichombo/sound/freq"
	"github.com/zikichombo/sound/gen"
	"github.com/zikichombo/sound/ops"
)

func TestPlay(t *testing.T) {
	src := ops.LimitDur(gen.Note(440*freq.Hertz), time.Second)
	start := time.Now()
	if err := sio.Play(src); err != nil {
		t.Fatal(err)
	}
	dur := time.Since(start)
	t.Logf("played for %s\n", dur)
}
