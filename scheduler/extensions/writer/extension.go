package writer

import (
	"io"
	"os"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions"
)

const TYPE = "writer"

type ExtensionDefinition struct {
	extensions.TypeBase[scheduler.ExtensionDefinition]
	writer io.Writer
}

var _ scheduler.ExtensionDefinition = (*ExtensionDefinition)(nil)

func Define(w io.Writer) scheduler.ExtensionDefinition {
	e := &ExtensionDefinition{
		writer: w,
	}
	e.TypeBase = extensions.NewBase[scheduler.ExtensionDefinition](e, TYPE, nil)
	return e
}

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	extensions.Extension
	scheduler scheduler.Scheduler
}

var _ scheduler.Extension = (*Extension)(nil)

func New(nested ...scheduler.Extension) scheduler.Extension {
	e := &Extension{}
	e.Extension = extensions.NewExtension(e, TYPE, nested...)
	return e
}

func (e *Extension) Setup(s scheduler.Scheduler) error {
	e.scheduler = s
	return e.Extension.Setup(s)
}

func (e *Extension) JobExtension(id string, jd scheduler.JobDefinition) (scheduler.JobExtension, error) {
	var err error

	j := &JobExtension{}
	j.JobExtension, err = extensions.NewJobExtension(j, TYPE, id, jd, e.Extension)
	if err != nil {
		return nil, err
	}

	def := generics.Cast[*ExtensionDefinition](jd.GetExtension(TYPE))
	if def != nil {
		j.writer = def.writer
	} else {
		j.writer = os.Stdout
	}

	return j, nil
}

////////////////////////////////////////////////////////////////////////////////

func GetExtension(job scheduler.Job) *JobExtension {
	return generics.Cast[*JobExtension](job.GetExtension(TYPE))
}

type JobExtension struct {
	extensions.JobExtension
	writer io.Writer
}

var _ scheduler.JobExtension = (*JobExtension)(nil)

func (j *JobExtension) Writer() io.Writer {
	return j.writer
}
