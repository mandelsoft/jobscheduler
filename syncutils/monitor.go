package syncutils

import (
	"context"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/syncutils/utils"
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
	entry   *utils.Entry[*blocker]
	waiting *waiting
}

func (b Block) Wait(ctx context.Context) error {
	err := b.entry.Elem.Wait(ctx)

	// remove entry.
	// if it is still there, there was no Unblock call for this
	// entry and a Block error has to be returned.
	// if the entry is already gone, there was an Unblock and a potential
	// interfering timeout can be ignored.
	p := &b.waiting.waiting
	for *p != nil {
		if *p == b.entry {
			*p = b.entry.Next
			return err
		}
		p = &(*p).Next
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Waiting provides basic monitor functionality.
// All operations are assumed to executed under a lock.
type Waiting interface {
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
	Prebook() Block

	// Wait registers the Go routine in a waiting list and blocks it until
	// the entry is deblocked again by a Deblock call.
	// If the optional sync.Locker is given the lock
	// is unlocked after the registration and before the go routine
	// is blocked.
	// If Wait is cancelled by the context an error is returned.
	// If the sync.Locker is given, the lock is in lock state
	// after this method returns.
	Wait(ctx context.Context, opt ...sync.Locker) error

	// Signal deblocks the first waiting go routine, if at least
	// one is blocked. It returns false if no one is found.
	// Signal must be called under the lock of the synchronized
	// data structure. The lock is transferred to the deblocked
	// go routine and true is returned.
	// If the optional sync.Locker is given, it is released if
	// no waiting go-routine could be found to transfer the lock to
	// and false is returned.
	Signal(opt ...sync.Locker) bool
}

type waiting struct {
	waiting *utils.Entry[*blocker]
}

func NewWaiting() Waiting {
	return &waiting{}
}

func (w *waiting) Prebook() Block {
	entry := &utils.Entry[*blocker]{Elem: newBlocker(), Next: w.waiting}
	w.waiting = entry
	return Block{entry, w}
}

func (w *waiting) Wait(ctx context.Context, opt ...sync.Locker) error {
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

func (w *waiting) Signal(opt ...sync.Locker) bool {
	l := general.Optional(opt...)
	if w.waiting == nil {
		log.Debug("nothing found to deblock")
		if l != nil {
			l.Unlock()
		}
		return false
	}

	p := &w.waiting
	for (*p).Next != nil {
		p = &(*p).Next
	}

	lock := (*p).Elem
	*p = nil
	lock.Signal()
	log.Debug("deblock succeeded")
	return true
}

////////////////////////////////////////////////////////////////////////////////

type Monitor interface {
	sync.Locker

	// Wait registers the Go routine in a waiting list and blocks it until
	// the entry is deblocked again by a Signal call.
	// It must be called under a lock, which is released by this call
	// before the go routine is finally blocked.
	// An error is returned, if the blocked go routine is canceled
	// by the context.
	// In any case the lock is locked after Wait returns.
	Wait(ctx context.Context) error

	// Signal deblocks the first waiting go routine, if at least
	// one is blocked. It returns false if no one is found.
	// Signal must be called under the monitor lock.
	// The lock is transferred to the deblocked
	// go routine and true is returned.
	// If no waiting go-routine could be found to transfer the lock to
	// the lock is released and false is returned.
	Signal() bool
}

type monitorOld struct {
	sync.Locker
	waiting *utils.Entry[*blocker]
}

func (w *monitorOld) Wait(ctx context.Context) error {
	entry := &utils.Entry[*blocker]{Elem: newBlocker(), Next: w.waiting}
	w.waiting = entry
	w.Unlock()
	err := entry.Elem.Wait(ctx)
	// remove entry.
	// if it is still there, there was no Unblock call for this
	// entry and a Block error has to be returned.
	// if the entry is already gone, there was an Unblock and a potential
	// interfering timeout can be ignored.
	p := &w.waiting
	for *p != nil {
		if *p == entry {
			*p = entry.Next
			if err != nil {
				w.Lock()
			}
			return err
		}
		p = &(*p).Next
	}
	return nil
}

func (w *monitorOld) Signal() bool {
	if w.waiting == nil {
		log.Debug("no one waiting")
		w.Unlock()
		return false
	}

	p := &w.waiting
	for (*p).Next != nil {
		p = &(*p).Next
	}

	lock := (*p).Elem
	*p = nil
	lock.Signal()
	log.Debug("deblock succeeded")
	return true
}

type monitor struct {
	sync.Locker
	waiting waiting
}

func NewMonitor(l sync.Locker) Monitor {
	return &monitor{Locker: l}
}

func NewMutexMonitor() Monitor {
	return &monitor{Locker: &sync.Mutex{}}
}

func (w *monitor) Wait(ctx context.Context) error {
	return w.waiting.Wait(ctx, w.Locker)
}

func (w *monitor) Signal() bool {
	return w.waiting.Signal(w.Locker)
}
