package queue

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/syncutils"
	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

type Priority int

func (p Priority) Less(other Priority) bool {
	return p > other
}

type Prioritized interface {
	GetPriority() Priority
}

// QueueElement is always a pointer to a struct
// implementing Prioritized.
// This is required to support == for removing
// elements.
type QueueElement[P any] interface {
	*P
	Prioritized
}

type Queue[E any, P QueueElement[E]] interface {
	Add(elem P)
	Remove(elem P)
	TryGet() (P, bool)
	Get(ctx context.Context) (P, error)
}

////////////////////////////////////////////////////////////////////////////////

type queue[E any, P QueueElement[E]] struct {
	monitor  syncutils.Monitor
	first    *utils.Entry[P]
	describe func(P) string
}

func New[E any, P QueueElement[E]](describe ...func(P) string) Queue[E, P] {
	return &queue[E, P]{monitor: syncutils.NewMutexMonitor(), describe: general.Optional(describe...)}
}

func (q *queue[E, P]) Add(elem P) {
	q.monitor.Lock()

	p := &q.first
	prio := elem.GetPriority()
	entry := &utils.Entry[P]{Elem: elem}

	for {
		if *p == nil || (*p).Elem.GetPriority().Less(prio) {
			entry.Next = *p
			*p = entry
			log.Debug("queue elem added")
			break
		}
		p = &(*p).Next
	}

	if !q.monitor.Signal() {
		log.Debug("nobody waiting for queue entry")
	}
}

func (q *queue[E, P]) Remove(elem P) {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	p := &q.first
	for *p != nil {
		if (*p).Elem == elem {
			*p = (*p).Next
		} else {
			p = &(*p).Next
		}
	}
}

func (q *queue[E, P]) TryGet() (P, bool) {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	if q.first == nil {
		return nil, false
	}
	elem := q.first.Elem
	q.first = q.first.Next
	return elem, true
}

func (q *queue[E, P]) Get(ctx context.Context) (P, error) {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	if q.first == nil {
		log.Debug("queue empty -> block")
		err := q.monitor.Wait(ctx)
		log.Debug("queue block deblocked", "error", err)
		if err != nil {
			return nil, err
		}
	}

	elem := q.first.Elem
	q.first = q.first.Next
	log.Debug("queue got element", "element", elem)
	return elem, nil
}
