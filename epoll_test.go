package goepoll

import (
	"net"
	"testing"
	"time"
)

func TestGoEpoll(t *testing.T) {
	tc, err := net.Dial("tcp", "127.0.0.1:9998")
	if err != nil {
		t.Error(err)
		return
	}
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
	if _, err := ev.Add(c, EVENT_READABLE); err != nil {
		t.Error(err)
		return
	}
	t.Log(c.Write([]byte("hello world")))
	t.Log(c.Write([]byte("hello world")))
	t.Log(c.Write([]byte("hello world")))
	<-time.After(time.Minute)
}
