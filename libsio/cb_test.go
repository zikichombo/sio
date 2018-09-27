package libsio

import (
	"fmt"
	"testing"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

func TestCb(t *testing.T) {
	N := 1024
	v := sound.MonoCd()
	c := sample.SFloat32L
	b := 512
	cb := NewCb(v, c, b)
	fmt.Printf("c cb addr from go %p\n", cb.c)
	go runcbs(cb, N, b, c.Bytes())
	d := make([]float64, b)
	for i := 0; i < N; i++ {
		fmt.Printf("cb receive %d\n", i)
		n, err := cb.Receive(d)
		if err != nil {
			t.Error(err)
		} else if n != b {
			t.Errorf("expected %d got %d\n", b, n)
		}
	}
}
