package goepoll

import (
	"net"
	"testing"
	"time"
)

func TestGoEpoll(t *testing.T) {
	tc, err := net.Dial("tcp", "tcpbin.com:4242")
	if err != nil {
		t.Error(err)
		return
	}
	defer tc.Close()
	t.Log(tc.Write([]byte("hello world")))
	go func() {
		buf := make([]byte, 100)
		n, err := tc.Read(buf)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(buf[:n]))
	}()
	onread := func(c *Conn) {
		t.Log("onread")
	}

	c := NewConn(tc, onread, nil, nil)
	ev, err := New()
	if err != nil {
		t.Error(err)
		return
	}
	defer ev.Close()
	if _, err := ev.Add(c, EVENT_READABLE); err != nil {
		t.Error(err)
		return
	}
	t.Log(c.Write([]byte("hello world")))
	<-time.After(time.Minute)
}
