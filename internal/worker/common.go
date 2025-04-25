package worker

import (
	"sync"
	"time"
)

type IdQueue struct {
	m            sync.Map
	debounceTime time.Duration
	transform    func(id string) string
}

func (q *IdQueue) Queue(sid string) {
	q.m.Swap(q.transform(sid), time.Now().Add(q.debounceTime))
}

func (q *IdQueue) delete(sid string) {
	q.m.Delete(q.transform(sid))
}

type WorkerQueueItem[T any] struct {
	v T
	t time.Time
}

type WorkerQueue[T any] struct {
	m            sync.Map
	getKey       func(item T) string
	transform    func(item *T) *T
	debounceTime time.Duration
}

func (q *WorkerQueue[T]) Queue(item T) {
	item = *q.transform(&item)
	q.m.Swap(q.getKey(item), WorkerQueueItem[T]{
		v: item,
		t: time.Now().Add(q.debounceTime),
	})
}

func (q *WorkerQueue[T]) delete(item T) {
	q.m.Delete(q.getKey(item))
}

func (q *WorkerQueue[T]) process(f func(item T)) {
	q.m.Range(func(k, v any) bool {
		_, keyOk := k.(string)
		val, valOk := v.(WorkerQueueItem[T])
		if keyOk && valOk && val.t.Before(time.Now()) {
			f(val.v)
			q.delete(val.v)
		}
		return true
	})
}
