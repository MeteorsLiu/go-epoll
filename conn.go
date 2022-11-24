package goepoll

import (
	"net"
	"time"

	"github.com/MeteorsLiu/mutex"
)

type Conn struct {
	conn net.Conn
	fd   int
	mu   struct {
		r mutex.Mutex
		w mutex.Mutex
		d mutex.Mutex
	}
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
	if c.onReadable != nil {
		return c.mu.r.TryLock()
	}
	return false
}

func (c *Conn) CouldBeWritable() bool {
	if c.onWritable != nil {
		return c.mu.w.TryLock()
	}
	return false
}

func (c *Conn) CouldBeDisconnected() bool {
	if c.onDisconnected != nil {
		return c.mu.d.TryLock()
	}
	return false
}
func (c *Conn) OnReadable() {
	if !c.mu.r.IsLocked() {
		return
	}
	defer c.mu.r.Unlock()
	c.onReadable(c)
}

func (c *Conn) OnWritable() {
	if !c.mu.w.IsLocked() {
		return
	}
	defer c.mu.w.Unlock()
	c.onWritable(c)
}

func (c *Conn) OnDisconnected() {
	if !c.mu.d.IsLocked() {
		return
	}
	defer c.mu.d.Unlock()
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
