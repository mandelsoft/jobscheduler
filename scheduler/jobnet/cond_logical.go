package jobnet

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

func prepareList(desc string, list []Condition, conds map[string]condition.Condition) error {
	result := errors.ErrList(desc)
	for _, e := range list {
		result.Add(e.Prepare(conds))
	}
	return result.Result()
}

func validateList(desc string, list []Condition, jobs map[string]Job) (set.Set[string], error) {
	deps := set.Set[string]{}

	result := errors.ErrList(desc)
	for _, e := range list {
		n, err := e.Validate(jobs)
		deps.AddAll(n)
		result.Add(err)
	}
	return deps, result.Result()
}

func createList(desc string, list []Condition, create func(...condition.Condition) condition.Condition, ctx *NetContext) (condition.Condition, error) {
	var cond []condition.Condition

	result := errors.ErrListf(desc)
	for _, e := range list {
		n, err := e.Create(ctx)
		if err != nil {
			result.Add(err)
		} else {
			cond = append(cond, n)
		}
	}
	if result.Result() != nil {
		return nil, result.Result()
	}
	return create(cond...), nil
}

////////////////////////////////////////////////////////////////////////////////

func Or(conditions ...Condition) Condition {
	return &_Or{conditions}
}

type _Or struct {
	conditions []Condition
}

func (c *_Or) Prepare(conds map[string]condition.Condition) error {
	return prepareList("or condition", c.conditions, conds)
}

func (c *_Or) Create(ctx *NetContext) (condition.Condition, error) {
	return createList("or condition", c.conditions, condition.Or, ctx)
}

func (c *_Or) Validate(jobs map[string]Job) (set.Set[string], error) {
	return validateList("or condition", c.conditions, jobs)
}

////////////////////////////////////////////////////////////////////////////////

func And(conditions ...Condition) Condition {
	return &_And{conditions}
}

type _And struct {
	conditions []Condition
}

func (c *_And) Prepare(conds map[string]condition.Condition) error {
	return prepareList("and condition", c.conditions, conds)
}

func (c *_And) Create(ctx *NetContext) (condition.Condition, error) {
	return createList("and condition", c.conditions, condition.And, ctx)
}

func (c *_And) Validate(jobs map[string]Job) (set.Set[string], error) {
	return validateList("and condition", c.conditions, jobs)
}

////////////////////////////////////////////////////////////////////////////////

func Not(c Condition) Condition {
	return &_Not{[]Condition{c}}
}

type _Not struct {
	conditions []Condition
}

func (c *_Not) Prepare(conds map[string]condition.Condition) error {
	return prepareList("not condition", c.conditions, conds)
}

func (c *_Not) Create(ctx *NetContext) (condition.Condition, error) {
	return createList("not condition", c.conditions, c.create, ctx)
}

func (_ *_Not) create(c ...condition.Condition) condition.Condition {
	if len(c) > 0 {
		return c[0]
	} else {
		return nil
	}
}

func (c *_Not) Validate(jobs map[string]Job) (set.Set[string], error) {
	return validateList("not condition", c.conditions, jobs)
}
