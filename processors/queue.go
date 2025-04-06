package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/queue"
	"github.com/mandelsoft/jobscheduler/syncutils"
)

// Queue is a queue supporting limitation of
// the number of active potential queue consumers.
// Using DiscardGet an additional
// consumer can be added by discarding one
// actual consuming Go routine. This means
//   - either a waiting consumer is discarded
//   - or a new Get request is discarded.
//
// It blocks until an existing consumer could be
// discarded.
// Discarding is indicated to a consumer by returning no
// error, but no entry, also, for a Get operation.
// It works on a pointer to a QueueElement.
type Queue[E any, P queue.QueueElement[E]] interface {
	Add(elem P)
	Remove(elem P)

	// Get returns an element from the queue.
	// It blocks until an element can be delivered,
	// the given context is cancelled, or a discard
	// of the consumer is requested.
	// It returns a nil element if the Go routine should
	// be discarded and an error is the context has
	// been cancelled.
	Get(ctx context.Context) (P, error)

	// DiscardGet requests one Get to be discarded.
	// It blocks until a matching Get request could be
	// discarded.
	DiscardGet(ctx context.Context) error

	HasDiscarded() bool
	HasWaiting() bool
}

type _queue[E any, P queue.QueueElement[E]] struct {
	queue   queue.SyncedQueue[E, P]
	limiter Limiter[P]
}

func NewQueue[E any, P queue.QueueElement[E]](describe ...func(P) string) (Queue[E, P], Limiter[P]) {
	q := queue.NewSynced[E, P](describe...)
	l := NewLimiter[P](q.Monitor(), NotFunc(q.List().IsEmpty),
		q.List().RemoveFirst)
	return &_queue[E, P]{queue: q, limiter: l}, l
}

func (q *_queue[E, P]) Monitor() syncutils.Monitor {
	return q.queue.Monitor()
}

func (q *_queue[E, P]) DiscardGet(ctx context.Context) error {
	return q.limiter.Discard(ctx)
}

func (q *_queue[E, P]) HasDiscarded() bool {
	return q.limiter.HasDiscarded()
}

func (q *_queue[E, P]) HasWaiting() bool {
	return q.queue.HasWaiting()
}

func (q *_queue[E, P]) Add(elem P) {
	q.queue.Add(elem)
}

func (q *_queue[E, P]) Remove(elem P) {
	q.queue.Remove(elem)
}

func (q *_queue[E, P]) Get(ctx context.Context) (P, error) {
	_, e, err := q.limiter.Request(ctx)
	return e, err
}
