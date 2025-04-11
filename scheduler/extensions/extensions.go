package extensions

import (
	"io"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/modern-go/reflect2"
)

type Provider[T any] interface {
	GetExtension(typ string) T
}

type TypeBase[T Provider[T]] struct {
	self   T
	nested T
	typ    string
}

func NewBase[T Provider[T]](self T, typ string, nested T) TypeBase[T] {
	return TypeBase[T]{self: self, typ: typ, nested: nested}
}

func (t *TypeBase[T]) GetExtension(typ string) T {
	var _nil T

	if t.typ == typ {
		return t.self
	}
	if reflect2.IsNil(t.nested) {
		return _nil
	}
	return t.nested.GetExtension(typ)
}

func (t *TypeBase[T]) Nested() T {
	return t.nested
}

func (t *TypeBase[T]) Self() T {
	return t.self
}

////////////////////////////////////////////////////////////////////////////////

func _GetExtension[E, T any](p Provider[E], typ string) T {
	return generics.Cast[T](p.GetExtension(typ))
}

////////////////////////////////////////////////////////////////////////////////

type ExtensionDefinition struct {
	TypeBase[scheduler.ExtensionDefinition]
}

var _ scheduler.ExtensionDefinition = (*ExtensionDefinition)(nil)

func NewExtensionDefinition(self scheduler.ExtensionDefinition, typ string, nested ...scheduler.ExtensionDefinition) ExtensionDefinition {
	return ExtensionDefinition{NewBase[scheduler.ExtensionDefinition](self, typ, general.Optional(nested...))}
}

func (e *ExtensionDefinition) SetSelf(self scheduler.ExtensionDefinition) {
	e.self = self
}

func GetExtensionDefinition[E scheduler.ExtensionDefinition](p Provider[scheduler.ExtensionDefinition], typ string) E {
	var _nil E
	if p == nil {
		return _nil
	}
	return _GetExtension[scheduler.ExtensionDefinition, E](p, typ)
}

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	TypeBase[scheduler.Extension]
}

var _ scheduler.Extension = (*Extension)(nil)

func NewExtension(self scheduler.Extension, typ string, nested ...scheduler.Extension) Extension {
	return Extension{NewBase[scheduler.Extension](self, typ, general.Optional(nested...))}
}

func (e *Extension) Setup(s scheduler.Scheduler) error {
	if e.nested != nil {
		return e.nested.Setup(s)
	}
	return nil
}

func (e *Extension) JobExtension(id string, def scheduler.JobDefinition, parent scheduler.Job) (scheduler.JobExtension, error) {
	if e.nested != nil {
		return e.nested.JobExtension(id, def, parent)
	}
	return nil, nil
}

func GetExtensionD[E scheduler.Extension](p Provider[scheduler.Extension], typ string) E {
	return _GetExtension[scheduler.Extension, E](p, typ)
}

////////////////////////////////////////////////////////////////////////////////

type JobExtension struct {
	TypeBase[scheduler.JobExtension]
}

var _ scheduler.JobExtension = (*JobExtension)(nil)

func NewJobExtension(self scheduler.JobExtension, typ string, id string, def scheduler.JobDefinition, e Extension) (JobExtension, error) {
	nested, err := e.JobExtension(id, def, nil)
	if err != nil {
		return JobExtension{}, err
	}
	return JobExtension{NewBase[scheduler.JobExtension](self, typ, nested)}, nil
}

func (e *JobExtension) Writer() io.Writer {
	if e.nested != nil {
		return e.nested.Writer()
	}
	return nil
}

func (e *JobExtension) Start() {
	if e.nested != nil {
		e.nested.Start()
	}
}

func (e *JobExtension) SetState(state scheduler.State) {
	if e.nested != nil {
		e.nested.SetState(state)
	}
}

func (e *JobExtension) Close() error {
	if e.nested != nil {
		return e.nested.Close()
	}
	return nil
}

func GetJobExtension[E scheduler.JobExtension](p Provider[scheduler.JobExtension], typ string) E {
	return _GetExtension[scheduler.JobExtension, E](p, typ)
}
