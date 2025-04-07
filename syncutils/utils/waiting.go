package utils

import (
	"context"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/matcher"
)

type blocker struct {
	lock   sync.Mutex
	locked bool
	c      chan struct{}
}

func newBlocker() *blocker {
	return &blocker{c: make(chan struct{}, 1)}
}

func (b *blocker) Wait(ctx context.Context) (bool, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			log.Debug("waiting: cancelled", "locked", b.locked, "p", b)
			// coordinate cancel and deblock
			if b.Signal(false) {
				return false, ctx.Err()
			}
		case <-b.c:
		}
	}
	<-b.c
	log.Debug("waiting: deblocked", "locked", b.locked, "p", b)
	return b.locked, nil
}

// Signal signals the blocker continue
// a pending or upcoming Wait.
// The signal indicated whether a lock is transferred
// or not.
func (b *blocker) Signal(locked bool) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	// coordinate cancel ad deblock
	select {
	case <-b.c:
		// if already cancelled prepend not to be signalled to avoid
		// lock transfer.
		log.Debug("waiting: already deblocked", "locked", locked, "p", b)
		return false
	default:
	}
	b.locked = locked
	log.Debug("waiting: deblocking", "locked", locked, "p", b)
	close(b.c)
	return true
}

////////////////////////////////////////////////////////////////////////////////

type Block struct {
	blocker *blocker
	waiting *Waiting
}

func (b *Block) Wait(ctx context.Context) (bool, error) {
	locked, err := b.blocker.Wait(ctx)

	// remove entry.
	// if it is still there, there was no Unblock call for this
	// entry and a Block error has to be returned.
	// if the entry is already gone, there was an Unblock and a potential
	// interfering timeout can be ignored.
	if b.waiting.waiting.Remove(matcher.Equals(b.blocker)) {
		return locked, err
	}
	return locked, nil
}

////////////////////////////////////////////////////////////////////////////////

// WaitingHandler describes optional
// resource handing operations when waiting for
// a condition using Waiting.
type WaitingHandler interface {
	Release(ctx context.Context)
	Alloc(ctx context.Context) error
}

// Waiting provides basic monitor functionality
// usable to implement synchronization objects.
// All methods must be called under a lock held
// for the synchronization object implementation.
type Waiting struct {
	waiting List[*blocker]
	handler WaitingHandler
}

func NewWaiting(handler ...WaitingHandler) Waiting {
	return Waiting{handler: general.Optional(handler...)}
}

// HasWaiting returns whether there are waiting Go routines.
func (w *Waiting) HasWaiting() bool {
	return !w.waiting.IsEmpty()
}

// Prebook registers an intended Wait call
// in the waiting list to be found by a Signal call.
// It will be found and deblocked
// even if the Go routine is not effectively finally blocked.
// In this case the Block.Wait call does not block the routine.
// The returned Block object MUST be used to
// call Block.Wait to finally block the Go routine.
// These separated calls can be used to release
// a locked data structure between both calls.
// If the Block is released in between, Block.Wait does
// not block and directly returns.
func (w *Waiting) Prebook() Block {
	b := newBlocker()
	w.waiting.Prepend(b)
	return Block{b, w}
}

// Wait registers the Go routine in a waiting list and blocks it until
// the entry is deblocked again by a Deblock call.
// If the optional sync.Locker is given the lock
// is unlocked after the registration and before the go routine
// is blocked.
// If Wait is cancelled by the context an error is returned.
// If the sync.Locker is given, the lock is in lock state
// after this method returns.
func (w *Waiting) Wait(ctx context.Context, opt ...sync.Locker) error {
	b := w.Prebook()
	l := general.Optional(opt...)
	if l != nil {
		l.Unlock()
	}
	if w.handler != nil {
		w.handler.Release(ctx)
	}
	locked, err := b.Wait(ctx)
	log.Debug("waiting: wait done", "locked", locked, "error", err)
	if !locked {
		if w.handler != nil {
			err2 := w.handler.Alloc(ctx)
			if err == nil {
				err = err2
			}
		}
		if l != nil {
			l.Lock()
		}
	}
	return err
}

// Signal deblocks the first waiting go routine, if at least
// one is blocked. It returns false if no one is found.
// Signal must be called under the lock of the synchronized
// data structure. The lock is transferred to the deblocked
// go routine and true is returned.
// If the optional sync.Locker is given, it is released if
// no waiting go-routine could be found to transfer the lock to
// and false is returned.
func (w *Waiting) Signal(ctx context.Context, opt ...sync.Locker) bool {
	l := general.Optional(opt...)
	if w.waiting.IsEmpty() {
		log.Debug("nothing found to deblock")
		if l != nil {
			l.Unlock()
		}
		if w.handler != nil {
			w.handler.Release(ctx)
		}
		return false
	}

	done := w.waiting.RemoveLast().Signal(true)
	log.Debug("deblock succeeded", "done", done)
	return done
}

// SignalAll unblocks all waiting
// Go routines.
func (w *Waiting) SignalAll() bool {
	if w.waiting.IsEmpty() {
		return false
	}
	for !w.waiting.IsEmpty() {
		w.waiting.RemoveFirst().Signal(false)
	}
	return true
}
