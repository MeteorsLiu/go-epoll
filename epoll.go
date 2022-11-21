package goepoll

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type EpollEvent int

const (
	EVENT_READABLE EpollEvent = iota
	EVENT_WRITABLE
	EVENT_DISCONNECTED
)

var (
	ErrEvents         = fmt.Errorf("invald events")
	ErrConn           = fmt.Errorf("invalid net.conn")
	ErrEventsExist    = fmt.Errorf("the event existed")
	ErrEventsNonExist = fmt.Errorf("the event doesn't exist")
	ErrEpollAdd       = fmt.Errorf("epoll_ctl_add fail")
	ErrEpollMod       = fmt.Errorf("epoll_ctl_mod fail")
	ErrEpollDel       = fmt.Errorf("epoll_ctl_del fail")
)

const (
	DEFAULT_EVENTS_SIZE = 1024
	EPOLLET             = 0x80000000
)

type Epoll struct {
	events     []syscall.EpollEvent
	events_len int64
	epollfd    int
	fds        sync.Map
	isClose    context.Context
	close      context.CancelFunc
}

func events(e EpollEvent) uint32 {
	switch e {
	case EVENT_DISCONNECTED:
		return syscall.EPOLLRDHUP
	case EVENT_READABLE:
		return syscall.EPOLLIN
	case EVENT_WRITABLE:
		return syscall.EPOLLOUT
	default:
		return 0
	}
}

func fd(c net.Conn) int {
	if t, ok := c.(*net.TCPConn); ok {
		f, _ := t.File()
		return int(f.Fd())
	}
	if i, ok := c.(*net.IPConn); ok {
		f, _ := i.File()
		return int(f.Fd())
	}
	if u, ok := c.(*net.UDPConn); ok {
		f, _ := u.File()
		return int(f.Fd())
	}
	if n, ok := c.(*net.UnixConn); ok {
		f, _ := n.File()
		return int(f.Fd())
	}
	return 0
}
func New() (*Epoll, error) {
	e := &Epoll{
		events: make([]syscall.EpollEvent, DEFAULT_EVENTS_SIZE),
	}
	epfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	e.epollfd = epfd
	e.isClose, e.close = context.WithCancel(context.Background())
	go e.daemon()
	return e, nil
}
func (e *Epoll) Close() {
	e.close()
	syscall.Close(e.epollfd)
}
func (e *Epoll) Add(c net.Conn, ev ...EpollEvent) (*Conn, error) {
	var cfd int
	var cn *Conn
	if cc, ok := c.(*Conn); ok {
		cfd = cc.Fd()
		cn = cc
	} else {
		cn = NewConn(c, nil, nil, nil)
		cfd = cc.Fd()
	}
	if cfd == 0 {
		return nil, ErrConn
	}
	if _, ok := e.fds.Load(cfd); ok {
		return nil, ErrEventsExist
	}
	if len(ev) == 0 || len(ev) > 3 {
		return nil, ErrEvents
	}
	evs := events(ev[0])
	for i := 1; i < len(ev); i++ {
		evs |= events(ev[i])
	}
	evs |= EPOLLET
	var event syscall.EpollEvent
	event.Events = evs
	event.Fd = int32(cfd)
	if err := syscall.EpollCtl(e.epollfd, syscall.EPOLL_CTL_ADD, cfd, &event); err != nil {
		return nil, ErrEpollAdd
	}
	log.Println("cfd: ", cfd)
	e.fds.Store(cfd, cn)
	atomic.AddInt64(&e.events_len, 1)
	return cn, nil
}

func (e *Epoll) Mod(c net.Conn, ev ...EpollEvent) (*Conn, error) {
	var cfd int
	var cn *Conn
	if cc, ok := c.(*Conn); ok {
		cfd = cc.Fd()
		cn = cc
	} else {
		cfd = fd(c)
		if cc, ok := e.fds.Load(cfd); !ok {
			return nil, ErrEventsNonExist
		} else {
			cn = cc.(*Conn)
		}
	}
	if cfd == 0 {
		return nil, ErrConn
	}

	if len(ev) == 0 || len(ev) > 3 {
		return nil, ErrEvents
	}
	evs := events(ev[0])
	for i := 1; i < len(ev); i++ {
		evs |= events(ev[i])
	}
	evs |= EPOLLET
	var event syscall.EpollEvent
	event.Events = evs
	event.Fd = int32(cfd)
	if err := syscall.EpollCtl(e.epollfd, syscall.EPOLL_CTL_MOD, cfd, &event); err != nil {
		return nil, ErrEpollMod
	}
	return cn, nil
}

func (e *Epoll) Del(c net.Conn) error {
	var cfd int
	if cc, ok := c.(*Conn); ok {
		cfd = cc.Fd()
	} else {
		cfd = fd(c)
	}
	if cfd == 0 {
		return ErrConn
	}
	if err := syscall.EpollCtl(e.epollfd, syscall.EPOLL_CTL_DEL, cfd, nil); err != nil {
		return ErrEpollDel
	}
	e.fds.Delete(cfd)
	atomic.AddInt64(&e.events_len, -1)
	return nil
}

func (e *Epoll) daemon() {
	for {
		size := atomic.LoadInt64(&e.events_len)
		if size == 0 {
			select {
			case <-e.isClose.Done():
				return
			default:
				time.Sleep(time.Second)
				continue
			}
		}
		if size > DEFAULT_EVENTS_SIZE {
			// resize if the number of connection is more than 1024
			e.events = make([]syscall.EpollEvent, size)
		}
		n, err := syscall.EpollWait(e.epollfd, e.events[:size], -1)
		if err != nil {
			select {
			case <-e.isClose.Done():
				return
			default:
				log.Println(err)
				continue
			}
		}
		for i := 0; i < n; i++ {
			if c, ok := e.fds.Load(e.events[i].Fd); ok {
				cn := c.(*Conn)
				if e.events[i].Events&(syscall.EPOLLERR|syscall.EPOLLRDHUP|syscall.EPOLLHUP) != 0 {
					if cn.HasDisconnector() {
						cn.OnDisconnected()
					}
				} else {
					if e.events[i].Events&syscall.EPOLLIN != 0 && cn.HasReader() {
						cn.OnReadable()
					}
					if e.events[i].Events&syscall.EPOLLOUT != 0 && cn.HasWriter() {
						cn.OnWritable()
					}
				}
			}
		}
	}
}
