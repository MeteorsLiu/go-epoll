package goepoll

import (
	"github.com/MeteorsLiu/go-epoll/worker"
)

type Queue struct {
	r *worker.Pool
	w *worker.Pool
	d *worker.Pool
}

func NewQueue(size int) *Queue {
	return &Queue{
		r: worker.NewPool(1, size, 1),
		w: worker.NewPool(1, size, 1),
		d: worker.NewPool(1, size, 1),
	}
}

func (q *Queue) Close() {
	q.r.Close()
	q.w.Close()
	q.d.Close()
}

func (q *Queue) ReadSchedule(f func()) {
	q.r.Schedule(f)
}

func (q *Queue) WriteSchedule(f func()) {
	q.w.Schedule(f)
}

func (q *Queue) DisconnectSchedule(f func()) {
	q.d.Schedule(f)
}
