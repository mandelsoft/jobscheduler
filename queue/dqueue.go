package queue

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

// DiscardQueue is a queue supporting limitation of
// the number of potential queue consumers.
// Using DiscardRequest an additional
// consumer can be added by discarding one
// actual consumers. This means
//   - either a waiting consumer is discarded
//   - or a new Get request is discarded.
//
// It blocks until a matching Get request could be
// discarded.
// Discarding is indicated by returning no
// error, but no entry, also.
type DiscardQueue[E any, P QueueElement[E]] interface {
	HasWaiting() bool
	HasDiscarded() bool

	Add(elem P)
	Remove(elem P)

	// Get returns an element from the queue.
	// It blocks until an element can be delivered,
	// the given context is cancelled, or a discard
	// is requested.
	Get(ctx context.Context) (P, error)

	// DiscardRequest requests one Get to be discarded.
	// It blocks until a matching Get request could be
	// discarded.
	DiscardRequest(ctx context.Context) error
}

type discardQueue[E any, P QueueElement[E]] struct {
	queue   queue[E, P]
	blocked int
	discard int
	waiting utils.Waiting
}

func NewDiscard[E any, P QueueElement[E]](describe ...func(P) string) DiscardQueue[E, P] {
	return &discardQueue[E, P]{queue: _New[E, P]()}
}

func (q *discardQueue[E, P]) HasWaiting() bool {
	q.queue.monitor.Lock()
	defer q.queue.monitor.Unlock()

	return q.queue.monitor.HasWaiting()
}

func (q *discardQueue[E, P]) HasDiscarded() bool {
	q.queue.monitor.Lock()
	defer q.queue.monitor.Unlock()

	return q.discard != 0
}

func (q *discardQueue[E, P]) DiscardRequest(ctx context.Context) error {
	q.queue.monitor.Lock()

	q.discard++
	if q.queue.monitor.HasWaiting() {
		q.queue.monitor.Signal()
		return nil
	}

	defer q.queue.monitor.Unlock()
	if err := q.waiting.Wait(ctx, q.queue.monitor); err != nil {
		return err
	}
	return nil
}

func (q *discardQueue[E, P]) Add(elem P) {
	q.queue.Add(elem)
}

func (q *discardQueue[E, P]) Remove(elem P) {
	q.queue.Remove(elem)
}

func (q *discardQueue[E, P]) Get(ctx context.Context) (P, error) {
	q.queue.monitor.Lock()

	if q.discard > 0 {
		q.discard--
		q.waiting.Signal(q.queue.monitor)
		return nil, nil
	}

	defer q.queue.monitor.Unlock()

	if q.queue.list.IsEmpty() {
		log.Debug("queue empty -> block")
		q.blocked++
		err := q.queue.monitor.Wait(ctx)
		q.blocked--
		log.Debug("queue block deblocked", "error", err)
		if err != nil {
			return nil, err
		}
		if q.discard > 0 {
			q.discard--
			return nil, nil
		}
	}

	elem, _ := q.queue.tryGet()
	log.Debug("queue got element", "element", elem)
	return elem, nil
}
