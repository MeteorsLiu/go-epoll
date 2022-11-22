package goepoll

import (
	"net"
	"sync"
	"time"
)

type Conn struct {
	conn           net.Conn
	fd             int
	rwMutex        sync.Mutex
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

func NewConnWithFd(c net.Conn, fd int, onread func(c *Conn), onwrite func(c *Conn), ondisconnect func(c *Conn)) *Conn {
	return &Conn{
		conn:           c,
		fd:             fd,
		onReadable:     onread,
		onWritable:     onwrite,
		onDisconnected: ondisconnect,
	}
}
func (c *Conn) Fd() int {
	return c.fd
}
func (c *Conn) HasReader() bool {
	return c.onReadable != nil
}

func (c *Conn) HasWriter() bool {
	return c.onWritable != nil
}

func (c *Conn) HasDisconnector() bool {
	return c.onDisconnected != nil
}
func (c *Conn) OnReadable() {
	// protect each R/W goroutine
	if !c.rwMutex.TryLock() {
		// A goroutine is still reading, don't interrupt
		// It happens that a goroutine try to read many times.
		// However, the event will be triggered twice.
		return
	}
	defer c.rwMutex.Unlock()
	c.onReadable(c)
}

func (c *Conn) OnWritable() {
	if !c.rwMutex.TryLock() {
		return
	}
	defer c.rwMutex.Unlock()
	c.onWritable(c)
}

func (c *Conn) OnDisconnected() {
	if !c.rwMutex.TryLock() {
		return
	}
	defer c.rwMutex.Unlock()
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
