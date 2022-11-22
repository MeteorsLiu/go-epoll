package goepoll

import (
	"net"
	"time"

	"github.com/MeteorsLiu/mutex"
)

type Conn struct {
	conn           net.Conn
	fd             int
	rwMutex        mutex.Mutex
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

// protect each R/W goroutine
// A goroutine is still reading, don't interrupt
// It happens that a goroutine try to read many times.
// However, the event will be triggered twice.
func (c *Conn) CouldBeReadable() bool {
	return c.onReadable != nil && c.rwMutex.TryLock()
}

func (c *Conn) CouldBeWritable() bool {
	return c.onWritable != nil && c.rwMutex.TryLock()
}

func (c *Conn) CouldBeDisconnected() bool {
	return c.onDisconnected != nil && c.rwMutex.TryLock()
}
func (c *Conn) OnReadable() {
	if !c.rwMutex.IsLocked() {
		return
	}
	defer c.rwMutex.Unlock()
	c.onReadable(c)
}

func (c *Conn) OnWritable() {
	if !c.rwMutex.IsLocked() {
		return
	}
	defer c.rwMutex.Unlock()
	c.onWritable(c)
}

func (c *Conn) OnDisconnected() {
	if !c.rwMutex.IsLocked() {
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
