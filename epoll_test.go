package goepoll

import (
	"net"
	"testing"
)

func TestGoEpoll(t *testing.T) {
	tc, _ := net.Dial("tcp", "tcpbin.com:4242")
	defer tc.Close()
	onread := func(c *Conn) {
		buf := make([]byte, 100)
		n, err := c.Read(buf)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(buf[:n]))
	}

	c := NewConn(tc, onread, nil, nil)
	ev, err := New()
	if err != nil {
		t.Error(err)
		return
	}
	defer ev.Close()
	ev.Add(c, EVENT_READABLE)
	c.Write([]byte("hello world"))
}
