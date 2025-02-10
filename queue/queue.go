package queue

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/matcher"
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
	HasWaiting() bool
}

////////////////////////////////////////////////////////////////////////////////

type queue[E any, P QueueElement[E]] struct {
	monitor  syncutils.Monitor
	list     utils.List[P]
	describe func(P) string
}

func New[E any, P QueueElement[E]](describe ...func(P) string) Queue[E, P] {
	return &queue[E, P]{monitor: syncutils.NewMutexMonitor(), describe: general.Optional(describe...)}
}

func _New[E any, P QueueElement[E]](describe ...func(P) string) queue[E, P] {
	return queue[E, P]{monitor: syncutils.NewMutexMonitor(), describe: general.Optional(describe...)}
}

func (q *queue[E, P]) HasWaiting() bool {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	return q.monitor.HasWaiting()
}

func (q *queue[E, P]) addToQueue(elem P) {
	prio := elem.GetPriority()
	q.list.Insert(elem, func(e P) bool {
		return e.GetPriority().Less(prio)
	})
}

func (q *queue[E, P]) removeFromQueue(elem P) {
	q.list.Remove(matcher.Equals(elem))
}

func (q *queue[E, P]) tryGet() (P, bool) {
	return q.list.RemoveFirst2()
}

func (q *queue[E, P]) Add(elem P) {
	q.monitor.Lock()
	q.addToQueue(elem)
	if !q.monitor.Signal() {
		log.Debug("nobody waiting for queue entry")
	}
}

func (q *queue[E, P]) Remove(elem P) {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	q.removeFromQueue(elem)
}

func (q *queue[E, P]) TryGet() (P, bool) {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	return q.tryGet()
}

func (q *queue[E, P]) Get(ctx context.Context) (P, error) {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	if q.list.IsEmpty() {
		log.Debug("queue empty -> block")
		err := q.monitor.Wait(ctx)
		log.Debug("queue block deblocked", "error", err)
		if err != nil {
			return nil, err
		}
	}

	elem := q.list.RemoveFirst()
	log.Debug("queue got element", "element", elem)
	return elem, nil
}
