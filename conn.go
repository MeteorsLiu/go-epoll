package goepoll

import (
	"net"
	"time"
)

type Conn struct {
	conn           net.Conn
	fd             int
	OnReadable     func()
	OnDisconnected func()
	OnWritable     func()
	ErrQueue       chan error
}

func NewConn(c net.Conn, fd int, onread func(), onwrite func(), ondisconnect func()) *Conn {
	return &Conn{
		conn:           c,
		fd:             fd,
		OnReadable:     onread,
		OnWritable:     onwrite,
		OnDisconnected: ondisconnect,
		ErrQueue:       make(chan error, 10),
	}
}
func (c *Conn) Fd() int {
	return c.fd
}

func (c *Conn) WithOnReadable(onread func()) {
	c.OnReadable = onread
}

func (c *Conn) WithOnWritable(onwrite func()) {
	c.OnWritable = onwrite
}
func (c *Conn) WithOnDisconnected(ondisconnect func()) {
	c.OnDisconnected = ondisconnect
}
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}
