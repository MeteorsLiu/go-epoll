package goepoll

import (
	"net"
	"time"
)

type Conn struct {
	conn           net.Conn
	fd             int
	onReadable     func(c *Conn)
	onDisconnected func(c *Conn)
	onWritable     func(c *Conn)
}

func NewConn(c net.Conn, onread func(c *Conn), onwrite func(c *Conn), ondisconnect func(c *Conn)) *Conn {
	return &Conn{
		conn:           c,
		fd:             fd(c),
		onReadable:     onread,
		onWritable:     onwrite,
		onDisconnected: ondisconnect,
	}
}
func (c *Conn) Fd() int {
	return c.fd
}

func (c *Conn) OnReadable() {
	c.onReadable(c)
}

func (c *Conn) OnWritable() {
	c.onWritable(c)
}

func (c *Conn) OnDisconnected() {
	c.onDisconnected(c)
}

func (c *Conn) WithOnReadable(onread func(c *Conn)) {
	c.onReadable = onread
}

func (c *Conn) WithOnWritable(onwrite func(c *Conn)) {
	c.onWritable = onwrite
}
func (c *Conn) WithOnDisconnected(ondisconnect func(c *Conn)) {
	c.onDisconnected = ondisconnect
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
