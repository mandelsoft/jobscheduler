package utils

import (
	"context"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/matcher"
)

type blocker struct {
	c chan struct{}
}

func newBlocker() *blocker {
	return &blocker{c: make(chan struct{}, 1)}
}

func (b *blocker) Wait(ctx context.Context) error {
	if ctx != nil {
		select {
		case <-b.c:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	} else {
		<-b.c
		return nil
	}
}

func (b *blocker) Signal() {
	// b.c <- struct{}{}
	close(b.c)
}

////////////////////////////////////////////////////////////////////////////////

type Block struct {
	blocker *blocker
	waiting *Waiting
}

func (b *Block) Wait(ctx context.Context) error {
	err := b.blocker.Wait(ctx)

	// remove entry.
	// if it is still there, there was no Unblock call for this
	// entry and a Block error has to be returned.
	// if the entry is already gone, there was an Unblock and a potential
	// interfering timeout can be ignored.
	if b.waiting.waiting.Remove(matcher.Equals(b.blocker)) {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Waiting provides basic monitor functionality
// usable to implement synchronization objects.
// All methods must be called under a lock held
// for the synchronization object implementation.
type Waiting struct {
	waiting List[*blocker]
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
	err := b.Wait(ctx)
	if err != nil {
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
func (w *Waiting) Signal(opt ...sync.Locker) bool {
	l := general.Optional(opt...)
	if w.waiting.IsEmpty() {
		log.Debug("nothing found to deblock")
		if l != nil {
			l.Unlock()
		}
		return false
	}

	w.waiting.RemoveLast().Signal()
	log.Debug("deblock succeeded")
	return true
}

// SignalAll unblocks all waiting
// Go routines.
func (w *Waiting) SignalAll() bool {
	if w.waiting.IsEmpty() {
		return false
	}
	for !w.waiting.IsEmpty() {
		w.waiting.RemoveFirst().Signal()
	}
	return true
}
