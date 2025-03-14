package uiblocks

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/mandelsoft/goutils/general"
)

const DefaultView = 10

// ESC is the ASCII code for escape character
const ESC = 27

// ErrClosedPipe is the error returned when trying to writer is not listening
var ErrClosedPipe = errors.New("uilive: read/write on closed pipe")

// FdWriter is a writer with a file descriptor.
type FdWriter interface {
	io.Writer
	Fd() uintptr
}

type UIBlock struct {
	titleline   string
	view        int
	blocks      *UIBlocks
	payload     any
	next        *UIBlock
	auto        bool
	gap         string
	followupGap string
	contentGap  string

	startline bool

	buf    bytes.Buffer
	closed bool
	done   chan struct{}

	final []byte
}

type block = UIBlock

func newBlock(w *UIBlocks, view ...int) *UIBlock {
	return &UIBlock{
		blocks:    w,
		startline: true,
		view:      general.OptionalDefaulted(DefaultView, view...),
		done:      make(chan struct{})}
}

func (w *UIBlock) UIBlocks() *UIBlocks {
	return w.blocks
}

func (w *UIBlock) SetTitleLine(s string) *UIBlock {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()

	w.titleline = s
	return w
}

func (w *UIBlock) SetFinal(data string) *UIBlock {
	w.final = []byte(data)
	return w
}

func (w *UIBlock) SetAuto(b ...bool) *UIBlock {
	w.auto = general.OptionalDefaultedBool(true, b...)
	return w
}

func (w *UIBlock) SetGap(gap string) *UIBlock {
	w.gap = gap
	if w.followupGap == "" {
		w.followupGap = gap
	}
	return w
}

func (w *UIBlock) SetFollowUpGap(gap string) *UIBlock {
	w.followupGap = gap
	return w
}

func (w *UIBlock) SetContentGap(gap string) *UIBlock {
	w.contentGap = gap
	return w
}

func (w *UIBlock) SetPayload(p any) *UIBlock {
	w.payload = p
	return w
}

func (w *UIBlock) Payload() any {
	return w.payload
}

func (w *UIBlock) SetNext(n *UIBlock) {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	w.next = n
}

func (w *UIBlock) Next() *UIBlock {
	w.blocks.lock.RLock()
	defer w.blocks.lock.RUnlock()
	return w.next
}

func (w *UIBlock) Reset() {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	w.startline = true
	w.buf.Reset()
}

// Write save the contents of buf to the writer b. The only errors returned are ones encountered while writing to the underlying buffer.
func (w *UIBlock) Write(buf []byte) (n int, err error) {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	if w.closed {
		return 0, os.ErrClosed
	}

	if strings.HasPrefix(string(buf), "doing") {
		w.buf.String()
	}
	contentgap := w.followupGap + w.contentGap
	gap := contentgap
	if w.buf.Len() == 0 && w.titleline == "" {
		gap = w.gap + w.contentGap
	}
	if gap != "" {
		for _, b := range buf {
			if b == '\n' {
				w.startline = true
				gap = contentgap
			} else {
				if w.startline {
					w.buf.Write([]byte(gap))
				}
				w.startline = false
			}
			w.buf.WriteByte(b)
		}
	} else {
		n, err = w.buf.Write(buf)
	}
	if w.auto {
		w.blocks.requestFlush()
	}
	return n, err
}

func (w *UIBlock) Close() error {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	if w.closed {
		return os.ErrClosed
	}
	w.closed = true
	close(w.done)
	return w.blocks.discardBlock()
}

func (w *UIBlock) IsClosed() bool {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	return w.closed
}

func (w *UIBlock) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-w.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *UIBlock) Flush() error {
	return w.blocks.Flush()
}

type lineinfo struct {
	start    int
	implicit int
}

func (w *UIBlock) flush(final bool) (int, error) {
	lines := 0
	titleline := 0
	newline := false
	data := w.buf.Bytes()
	if w.closed && w.final != nil {
		data = w.final
	} else {
		if w.titleline != "" {
			w.blocks.out.Write([]byte(w.gap + w.titleline + "\n"))
			titleline = 1
		}
	}
	if len(data) == 0 {
		return titleline, nil
	}

	implicit := 0
	linestart := make([]lineinfo, w.view)

	escapeSequence := 0

	var col int
	start := 0
	// fmt.Fprintf(os.Stderr, "write [%d] %q\n", len(data), string(data))
	for o, b := range string(data) {
		if escapeSequence == 0 {
			escapeSequence = ColorLength(data[o:])
		}
		if escapeSequence == 0 && b == '\n' || (w.blocks.overFlowHandled && col >= w.blocks.termWidth) {
			if b != '\n' {
				implicit++
			} else {
				linestart[lines%w.view].start = start
				linestart[lines%w.view].implicit = implicit
				start = o + 1
				lines++
			}
			newline = true
			col = 0
		} else {
			// fmt.Fprintf(os.Stderr, "insert linebreak %d\n", col)
			newline = false
			if escapeSequence > 0 {
				escapeSequence--
			} else {
				col++
			}
		}
	}

	if !newline {
		linestart[lines%w.view].start = start
		linestart[lines%w.view].implicit = implicit
		lines++
		data = append(data, '\n')
	}

	if w.view > 1 {
		newline = false
	}

	var err error
	if final || lines <= w.view {
		_, err = w.blocks.out.Write(data)
		eff := lines + implicit + titleline
		// fmt.Fprintf(os.Stderr, "data: %s\n", string(data))
		// fmt.Fprintf(os.Stderr, "eff %d, lines %d, implicit %d\n", eff, lines, implicit)

		return eff, err
	} else {
		index := (lines) % w.view
		start := linestart[index].start
		view := data[start:]
		_, err = w.blocks.out.Write(view)
		eff := w.view + implicit - linestart[index].implicit + titleline
		// fmt.Fprintf(os.Stderr, "data: %s\n", string(view))
		// fmt.Fprintf(os.Stderr, "eff %d, lines %d, implicit %d\n", eff, lines, implicit)

		return eff, err
	}
}
