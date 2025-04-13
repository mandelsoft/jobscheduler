package scheduler

import (
	"io"
)

const defaultType = "default"

type _Extension struct {
	scheduler Scheduler
}

var _ Extension = (*_Extension)(nil)

func newDefaultExtension() Extension {
	return &_Extension{}
}

func (d *_Extension) GetExtension(typ string) Extension {
	if typ == defaultType {
		return d
	}
	return nil
}

func (e *_Extension) Setup(s Scheduler) error {
	e.scheduler = s
	return nil
}

func (e *_Extension) JobExtension(id string, definition JobDefinition, parent Job) (JobExtension, error) {
	return &_JobExtension{
		writer: Stdout,
	}, nil
}

func (e *_Extension) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type _JobExtension struct {
	writer io.Writer
}

var _ JobExtension = (*_JobExtension)(nil)

func (j *_JobExtension) GetExtension(typ string) JobExtension {
	if typ == defaultType {
		return j
	}
	return nil
}

func (j *_JobExtension) SetState(state State) {
}

func (j *_JobExtension) Writer() io.Writer {
	return j.writer
}

func (j *_JobExtension) Close() error {
	return nil
}

func (j *_JobExtension) Start() {
}
