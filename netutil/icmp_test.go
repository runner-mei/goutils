package netutil

import (
	"testing"
	"time"
)

func TestICMP2(t *testing.T) {
	icmp, err := NewPinger("ip4:icmp", "", []byte("gogogogogogogogo"), 100)
	if nil != err {
		t.Error(err)
		return
	}
	err = icmp.Send("127.0.0.1", nil)
	if nil != err {
		t.Error(err)
		return
	}
	ra, _, err := icmp.Recv(1 * time.Second)
	if nil != err {
		t.Error(err)
		return
	}

	t.Log(ra)

	err = icmp.Send("127.0.0.1", nil)
	if nil != err {
		t.Error(err)
		return
	}
	ra, _, err = icmp.Recv(1 * time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(ra)
}
