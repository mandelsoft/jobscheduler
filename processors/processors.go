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

func NewProcessors[E any](ctx context.Context, creator Creator, limiter Limiter[E], n ...int) *Processors[E] {
	ctx, cancel := context.WithCancel(ctx)
	p := &Processors[E]{
		limiter: limiter,
		creator: creator,
		ctx:     ctx,
		cancel:  cancel,
		runners: map[int]Runner{},
	}
	for i := 0; i < general.Optional(n...); i++ {
		p.New()
	}
	return p
}

func (p *Processors[E]) New() {
	p.lock.Lock()
	defer p.lock.Unlock()

	id := p.ids.Next()
	p.done.Add(1)
	r := p.creator(id)
	p.runners[id] = r
	go func() {
		r.Run(p.ctx)
		p.done.Done()
	}()
}

func (p *Processors[E]) DiscardRequest(ctx context.Context) error {
	return p.limiter.DiscardRequest(ctx)
}

func (p *Processors[E]) Wait() {
	p.done.Wait()
}
