package processors

import (
	"context"
	"slices"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/matcher"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/jobscheduler/ctxutils"
	"github.com/mandelsoft/jobscheduler/syncutils/synclog"
)

type Ids []int

func (ids *Ids) Next() int {
	last := 0
	for _, id := range *ids {
		if id > last+1 {
			break
		}
		last = id
	}
	last++
	*ids = sliceutils.InsertAscending(*ids, last)
	return last
}

func (ids *Ids) Remove(id int) {
	*ids = slices.DeleteFunc(*ids, matcher.Equals(id))
}

////////////////////////////////////////////////////////////////////////////////

var runnerAttr = ctxutils.NewAttribute[Runner]()

func GetRunner(ctx context.Context) Runner {
	return runnerAttr.Get(ctx)
}

func setRunner(ctx context.Context, r Runner) context.Context {
	return runnerAttr.Set(ctx, r)
}

////////////////////////////////////////////////////////////////////////////////

type Runner interface {
	Run(context.Context)
}

type Creator func(id int) Runner

type StateHandler interface {
	Ready(ctx context.Context)
	Running(ctx context.Context)
	Block(ctx context.Context)
}

type Processors[E any] struct {
	lock    synclog.Mutex
	limiter Limiter[E]
	creator Creator
	handler StateHandler

	ids     Ids
	runners map[int]Runner
	done    sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

var _ Pool = (*Processors[int])(nil)
var _ PoolProvider = (*Processors[int])(nil)

func NewProcessors[E any](creator Creator, limiter Limiter[E], n ...int) *Processors[E] {
	p := &Processors[E]{
		lock:    synclog.NewMutex("processors"),
		limiter: limiter,
		creator: creator,
		runners: map[int]Runner{},
		handler: &dummyHandler{},
	}
	for i := 0; i < general.Optional(n...); i++ {
		p.New()
	}
	return p
}

func (p *Processors[E]) GetPool() Pool {
	return p
}

func (p *Processors[E]) IsStarted() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.ctx != nil
}

func (p *Processors[E]) SetStateHandler(h StateHandler) {
	p.handler = h
}

func (p *Processors[E]) Cancel() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.ctx != nil {
		p.cancel()
	}
}

func (p *Processors[E]) Run(ctx context.Context) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.ctx != nil {
		return ErrAlreadyStarted
	}
	if ctx == nil {
		ctx = context.Background()
	}
	p.ctx, p.cancel = context.WithCancel(WithPool(ctx, p))

	for _, r := range p.runners {
		go func() {
			r.Run(setRunner(p.ctx, r))
			p.done.Done()
		}()
	}
	return nil
}

func (p *Processors[E]) New() {
	p.lock.Lock()
	defer p.lock.Unlock()

	id := p.ids.Next()
	p.done.Add(1)
	r := p.creator(id)
	p.runners[id] = r

	if p.ctx != nil {
		go func() {
			r.Run(p.ctx)
			p.done.Done()
		}()
	}
}

func (p *Processors[E]) Discard(ctx context.Context) error {
	return p.limiter.Discard(ctx)
}

func (p *Processors[E]) Wait() {
	p.done.Wait()
}

////////////////////////////////////////////////////////////////////////////////

func (p *Processors[E]) Alloc(ctx context.Context) error {
	p.handler.Ready(ctx)
	err := p.Discard(ctx)
	if err == nil {
		p.handler.Running(ctx)
	}
	return err
}

func (p *Processors[E]) Release(ctx context.Context) {
	p.handler.Block(ctx)
	p.New()
}

////////////////////////////////////////////////////////////////////////////////

type dummyHandler struct{}

var _ StateHandler = (*dummyHandler)(nil)

func (d *dummyHandler) Ready(ctx context.Context) {
}

func (d *dummyHandler) Running(ctx context.Context) {
}

func (d *dummyHandler) Block(ctx context.Context) {
}
