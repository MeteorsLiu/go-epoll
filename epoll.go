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

	"github.com/MeteorsLiu/go-epoll/worker"
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
	maxSize    int64
	events_len int64
	epollfd    int
	fds        sync.Map
	once       sync.Once
	wpool      *worker.Pool
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
func New(events_num ...int) (*Epoll, error) {
	size := DEFAULT_EVENTS_SIZE
	if len(events_num) > 0 {
		size = events_num[0]
	}
	epfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	e := &Epoll{
		events:  make([]syscall.EpollEvent, size),
		maxSize: int64(size),
		wpool:   worker.NewPool(size, size, size),
		epollfd: epfd,
	}
	e.isClose, e.close = context.WithCancel(context.Background())
	go e.daemon()
	return e, nil
}
func (e *Epoll) Close() {
	e.once.Do(func() {
		e.close()
		syscall.Close(e.epollfd)
	})
}
func (e *Epoll) Add(c net.Conn, ev ...EpollEvent) (*Conn, error) {
	var cfd int32
	var cn *Conn
	if len(ev) == 0 || len(ev) > 3 {
		return nil, ErrEvents
	}
	if cc, ok := c.(*Conn); ok {
		cfd = int32(cc.Fd())
		cn = cc
	} else {
		cn = NewConn(c, nil, nil, nil)
		cfd = int32(cc.Fd())
	}
	if cfd == 0 {
		return nil, ErrConn
	}
	if _, ok := e.fds.LoadOrStore(cfd, cn); ok {
		return nil, ErrEventsExist
	}
	evs := events(ev[0])
	for i := 1; i < len(ev); i++ {
		evs |= events(ev[i])
	}
	evs |= EPOLLET
	var event syscall.EpollEvent
	event.Events = evs
	event.Fd = cfd
	if err := syscall.EpollCtl(e.epollfd, syscall.EPOLL_CTL_ADD, int(cfd), &event); err != nil {
		return nil, ErrEpollAdd
	}
	atomic.AddInt64(&e.events_len, 1)
	return cn, nil
}

func (e *Epoll) Mod(c net.Conn, ev ...EpollEvent) (*Conn, error) {
	var cfd int32
	var cn *Conn
	if len(ev) == 0 || len(ev) > 3 {
		return nil, ErrEvents
	}
	if cc, ok := c.(*Conn); ok {
		cfd = int32(cc.Fd())
	} else {
		cfd = int32(fd(c))
	}
	if cfd == 0 {
		return nil, ErrConn
	}
	if cc, ok := e.fds.Load(cfd); !ok {
		return nil, ErrEventsNonExist
	} else {
		cn = cc.(*Conn)
	}
	evs := events(ev[0])
	for i := 1; i < len(ev); i++ {
		evs |= events(ev[i])
	}
	evs |= EPOLLET
	var event syscall.EpollEvent
	event.Events = evs
	event.Fd = cfd
	if err := syscall.EpollCtl(e.epollfd, syscall.EPOLL_CTL_MOD, int(cfd), &event); err != nil {
		return nil, ErrEpollMod
	}
	return cn, nil
}

func (e *Epoll) Del(c net.Conn) error {
	var cfd int32
	if cc, ok := c.(*Conn); ok {
		cfd = int32(cc.Fd())
	} else {
		cfd = int32(fd(c))
	}
	if cfd == 0 {
		return ErrConn
	}
	if _, ok := e.fds.Load(cfd); !ok {
		return ErrEventsNonExist
	}
	if err := syscall.EpollCtl(e.epollfd, syscall.EPOLL_CTL_DEL, int(cfd), nil); err != nil {
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
		if size > e.maxSize {
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
						e.wpool.Schedule(cn.OnDisconnected)
					}
				} else {
					if e.events[i].Events&syscall.EPOLLIN != 0 && cn.HasReader() {
						e.wpool.Schedule(cn.OnReadable)
					}
					if e.events[i].Events&syscall.EPOLLOUT != 0 && cn.HasWriter() {
						e.wpool.Schedule(cn.OnWritable)
					}
				}
			}
		}
	}
}
