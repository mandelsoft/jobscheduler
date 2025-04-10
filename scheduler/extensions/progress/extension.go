package progress

import (
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/jobscheduler/ctxutils"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions"
	"github.com/mandelsoft/ttyprogress"
)

const TYPE = "progress"

const (
	VAR_JOBSTATE = "jobstate"
	VAR_JOBID    = "jobid"
	VAR_JOBNAME  = "jobname"
)

type ExtensionDefinition struct {
	extensions.ExtensionDefinition
	hideOnClose bool
	progress    ttyprogress.ElementDefinition[ttyprogress.Element]
}

var _ scheduler.ExtensionDefinition = (*ExtensionDefinition)(nil)

func Define[T ttyprogress.ElementDefinition[E], E ttyprogress.ProgressElement](d T, nested ...scheduler.ExtensionDefinition) *ExtensionDefinition {
	e := &ExtensionDefinition{
		progress: ttyprogress.GenericDefinition(d),
	}
	e.ExtensionDefinition = extensions.NewExtensionDefinition(e, TYPE, nested...)
	return e
}

func (d *ExtensionDefinition) Dup() *ExtensionDefinition {
	e := *d
	e.SetSelf(&e)
	return &e
}

// HideOutputOnClose is only used, if the configured indicator does not
// act as writer (for example a ttyprogress.TextSpinner).
func (d *ExtensionDefinition) HideOutputOnClose(b ...bool) *ExtensionDefinition {
	c := d.Dup()
	c.hideOnClose = optionutils.BoolOption(b...)
	return c
}

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	extensions.Extension
	scheduler    scheduler.Scheduler
	pctx         ttyprogress.Context
	defaultGroup *ttyprogress.AnonymousGroupDefinition
}

var _ scheduler.Extension = (*Extension)(nil)

func New(p ttyprogress.Context, nested ...scheduler.Extension) scheduler.Extension {
	e := &Extension{
		pctx:         p,
		defaultGroup: ttyprogress.NewAnonymousGroup().SetGap("  "),
	}
	e.Extension = extensions.NewExtension(e, TYPE, nested...)
	return e
}

func (e *Extension) Setup(s scheduler.Scheduler) error {
	e.scheduler = s
	return e.Extension.Setup(s)
}

func (e *Extension) JobExtension(jid string, jd scheduler.JobDefinition, parent scheduler.Job) (scheduler.JobExtension, error) {
	var err error

	var ctx ttyprogress.Container = e.pctx

	if parent != nil {
		p := extensions.GetJobExtension[*JobExtension](parent, TYPE)
		if p != nil {
			ctx = p.group
		}
	}

	j := &JobExtension{}
	j.JobExtension, err = extensions.NewJobExtension(j, TYPE, jid, jd, e.Extension)
	if err != nil {
		return nil, err
	}

	def := extensions.GetExtensionDefinition[*ExtensionDefinition](jd.GetExtension(), TYPE)
	if def != nil {
		p, err := def.progress.Add(ctx)
		if err != nil {
			return nil, err
		}
		j.progress = generics.Cast[ttyprogress.ProgressElement](p)
		j.progress.SetVariable(VAR_JOBID, jid)
		j.progress.SetVariable(VAR_JOBNAME, jd.GetName())
		if w, ok := j.progress.(io.WriteCloser); ok {
			j.writer = w
		} else {
			j.writer, err = ttyprogress.NewText(3).
				SetGap("  ").
				HideOnClose(def.hideOnClose).
				Add(ctx)
			if err != nil {
				return nil, err
			}
		}
	} else {
		j.writer = ctxutils.NopCloser(os.Stdout)
	}

	j.group, err = e.defaultGroup.Add(ctx)
	if err != nil {
		return nil, err
	}

	return j, nil
}

////////////////////////////////////////////////////////////////////////////////

func GetExtension(job scheduler.Job) *JobExtension {
	return generics.Cast[*JobExtension](job.GetExtension(TYPE))
}

type JobExtension struct {
	extensions.JobExtension
	progress ttyprogress.ProgressElement
	group    ttyprogress.AnonymousGroup
	writer   io.WriteCloser
}

var _ scheduler.JobExtension = (*JobExtension)(nil)

func (j *JobExtension) GetIndicator() ttyprogress.ProgressElement {
	return j.progress
}

func (j *JobExtension) Writer() io.Writer {
	return j.writer
}

func (j *JobExtension) Close() error {
	var err errors.ErrorList
	if j.progress != nil {
		err.Add(j.progress.Close())
	}
	err.Add(j.writer.Close()).
		Add(j.group.Close()).
		Add(j.JobExtension.Close())
	return err.Result()
}

func (j *JobExtension) Start() {
	if j.progress != nil {
		j.progress.Start()
	}
	j.JobExtension.Start()
}

func (j *JobExtension) SetState(state scheduler.State) {
	if j.progress != nil {
		j.progress.SetVariable(VAR_JOBSTATE, state)
	}
	j.JobExtension.SetState(state)
}
