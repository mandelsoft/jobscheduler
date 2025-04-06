package processors

import (
	"context"
	"slices"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/matcher"
	"github.com/mandelsoft/goutils/sliceutils"
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

type Runner interface {
	Run(context.Context)
}

type Creator func(id int) Runner

type Processors[E any] struct {
	lock    sync.Mutex
	limiter Limiter[E]
	creator Creator

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
		limiter: limiter,
		creator: creator,
		runners: map[int]Runner{},
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
	p.ctx, p.cancel = context.WithCancel(ctx)

	for _, r := range p.runners {
		go func() {
			r.Run(p.ctx)
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
	return p.Discard(ctx)
}

func (p *Processors[E]) Release() {
	p.New()
}
