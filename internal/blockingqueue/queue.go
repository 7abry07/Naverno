package blockingqueue

import (
	"sync"
	"sync/atomic"
)

type Queue[T any] struct {
	mut   sync.Mutex
	cond  *sync.Cond
	stop  atomic.Bool
	items []T
}

func New[T any](items []T) *Queue[T] {
	q := &Queue[T]{items: items}
	q.cond = sync.NewCond(&q.mut)
	return q
}

func (q *Queue[T]) Stop() {
	q.mut.Lock()
	defer q.mut.Unlock()
	q.stop.Store(true)
	q.cond.Broadcast()
}

func (q *Queue[T]) Push(v T) {
	q.mut.Lock()
	defer q.mut.Unlock()
	q.items = append(q.items, v)
	q.cond.Signal()
}

func (q *Queue[T]) Pop() (T, bool) {
	q.mut.Lock()
	defer q.mut.Unlock()

	for len(q.items) == 0 && !q.stop.Load() {
		q.cond.Wait()
	}

	if q.stop.Load() && len(q.items) == 0 {
		var z T
		return z, false
	}

	v := q.items[0]
	q.items = q.items[1:]
	return v, true
}
