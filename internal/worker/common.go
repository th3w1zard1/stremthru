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
