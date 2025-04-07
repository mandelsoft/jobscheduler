package processors

import (
	"context"

	"github.com/mandelsoft/jobscheduler/syncutils"
	"github.com/mandelsoft/jobscheduler/syncutils/utils"
)

// Limiter handles the limitation of the parallel
// execution. It is based on a syncutils.Monitor
// a check and a consumption handling.
// The synchronization is completely handled
// using the given monitor.
type Limiter[E any] interface {
	// Request executes a consumption request, which
	// returns an E element.
	// The boolean indicates whether the request
	// hast been fulfilled (true) or should be discarded (false).
	Request(ctx context.Context) (bool, E, error)

	// Discard requests one consumption request to be discarded.
	// It blocks until a matching consumption request could be
	// discarded. Afterward, a new Go routine can continue replacing
	// the discarded one.
	Discard(ctx context.Context) error

	HasDiscarded() bool
}

type limiter[E any] struct {
	monitor syncutils.Monitor
	blocked int
	discard int
	// waiting holds go routines waiting to get the
	// permission to run until another routine has been
	// notified to be cancelled.
	waiting utils.Waiting

	// checker checks whether consumption is possible.
	checker func() bool
	// consumer finally consumes an element.
	consumer func() E
}

func NewLimiter[E any](m syncutils.Monitor, checker func() bool, consumer func() E) Limiter[E] {
	return &limiter[E]{monitor: m, checker: checker, consumer: consumer}
}

func (l *limiter[E]) Monitor() syncutils.Monitor {
	return l.monitor
}

func (q *limiter[E]) HasWaiting() bool {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	return q.monitor.HasWaiting()
}

func (q *limiter[E]) HasDiscarded() bool {
	q.monitor.Lock()
	defer q.monitor.Unlock()

	return q.discard != 0
}

func (q *limiter[E]) Discard(ctx context.Context) error {
	q.monitor.Lock()

	q.discard++
	log.Debug("discarding processor", "discarded", q.discard)
	if q.monitor.HasWaiting() {
		// wakeup waiting go routine to be cancelled
		log.Debug("signal waiting routine to be discarded")
		q.monitor.Signal(ctx)
		return nil
	}

	defer q.monitor.Unlock()
	log.Debug("wait for discarded routine")
	if err := q.waiting.Wait(ctx, q.monitor); err != nil {
		log.Debug("waiting aborted", "error", err)
		return err
	}
	log.Debug("routine was successfully discarded")
	return nil
}

func (q *limiter[E]) Request(ctx context.Context) (bool, E, error) {
	var _nil E

	q.monitor.Lock()

	if q.discard > 0 {
		q.discard--
		q.waiting.Signal(ctx, q.monitor)
		return false, _nil, nil
	}

	defer q.monitor.Unlock()

	if avail := q.checker(); !avail {
		q.blocked++
		err := q.monitor.Wait(ctx)
		q.blocked--
		if err != nil {
			log.Debug("wait aborted", "error", err)
			return false, _nil, err
		}
		if q.discard > 0 {
			q.discard--
			return false, _nil, nil
		}
	}
	return true, q.consumer(), nil
}
