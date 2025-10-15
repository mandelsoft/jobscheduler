package buffered

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/jobscheduler/scheduler"
	"github.com/mandelsoft/jobscheduler/scheduler/extensions"
)

const TYPE = "writer"

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	extensions.Extension
	lock      sync.Mutex
	scheduler scheduler.Scheduler
	writer    io.Writer
	blocks    []*JobExtension
}

var _ scheduler.Extension = (*Extension)(nil)

func New(writer io.Writer, nested ...scheduler.Extension) scheduler.Extension {
	e := &Extension{writer: writer}
	e.Extension = extensions.NewExtension(e, TYPE, nested...)
	return e
}

func (e *Extension) Setup(s scheduler.Scheduler) error {
	e.scheduler = s
	return e.Extension.Setup(s)
}

func (e *Extension) JobExtension(id string, jd scheduler.JobDefinition, parent scheduler.Job) (scheduler.JobExtension, error) {
	var err error

	e.lock.Lock()
	defer e.lock.Unlock()

	gap := ""
	if parent != nil {
		p := extensions.GetJobExtension[*JobExtension](parent, TYPE)
		if p != nil {
			gap = p.gap + "  "
		}
	}

	j := &JobExtension{
		ext:    e,
		id:     id,
		gap:    gap,
		writer: bytes.NewBuffer(nil),
	}
	j.JobExtension, err = extensions.NewJobExtension(j, TYPE, id, jd, e.Extension)
	if err != nil {
		return nil, err
	}
	e.blocks = append(e.blocks, j)
	return j, nil
}

func (e *Extension) discard() {
	e.lock.Lock()
	defer e.lock.Unlock()

	for len(e.blocks) > 0 && e.blocks[0].done.Load() {
		e.blocks[0].emit()
		e.blocks = e.blocks[1:]
	}
}

////////////////////////////////////////////////////////////////////////////////

func GetExtension(job scheduler.Job) *JobExtension {
	return generics.Cast[*JobExtension](job.GetExtension(TYPE))
}

type JobExtension struct {
	extensions.JobExtension
	ext    *Extension
	gap    string
	id     string
	writer *bytes.Buffer
	state  scheduler.State
	done   atomic.Bool
}

var _ scheduler.JobExtension = (*JobExtension)(nil)

func (j *JobExtension) Writer() io.Writer {
	return j.writer
}

func (j *JobExtension) SetState(state scheduler.State) {
	j.state = state
	if scheduler.IsFinished(state) {
		j.done.Store(true)
		j.ext.discard()
	}
}

func (j *JobExtension) emit() {
	fmt.Fprintf(j.ext.writer, "%s- JOB %s %s\n", j.gap, j.id, j.state)
	s := j.writer.String()
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	fmt.Fprint(j.ext.writer, j.gap+"  "+strings.ReplaceAll(s, "\n", "\n"+j.gap+"  ")+"\n")
}
