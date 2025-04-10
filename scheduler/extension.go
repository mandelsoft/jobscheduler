package scheduler

import (
	"io"
)

type Extension interface {
	GetExtension(typ string) Extension

	Setup(s Scheduler) error
	JobExtension(id string, definition JobDefinition, parent Job) (JobExtension, error)
}

type ExtensionDefinition interface {
	GetExtension(typ string) ExtensionDefinition
}

type JobExtension interface {
	GetExtension(typ string) JobExtension

	Writer() io.Writer
	Start()
	SetState(state State)
	Close() error
}
