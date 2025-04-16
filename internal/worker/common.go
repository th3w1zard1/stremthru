package worker

import (
	"sync"
	"time"
)

type IdQueue struct {
	m            sync.Map
	debounceTime time.Duration
}

func (q *IdQueue) Queue(sid string) {
	q.m.Swap(sid, time.Now().Add(q.debounceTime))
}

func (q *IdQueue) delete(sid string) {
	q.m.Delete(sid)
}
