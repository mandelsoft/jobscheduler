package uiblocks

import (
	"bytes"
	"errors"
	"io"
	"os"

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

type Block struct {
	titleline string
	view      int
	blocks    *UIBlocks
	payload   any
	auto      bool

	buf    bytes.Buffer
	closed bool

	final []byte
}

type block = Block

func (w *Block) UIBlocks() *UIBlocks {
	return w.blocks
}

func (w *Block) SetTitleLine(s string) *Block {
	w.titleline = s
	return w
}

func (w *Block) SetFinal(data string) *Block {
	w.final = []byte(data)
	return w
}

func (w *Block) SetAuto(b ...bool) *Block {
	w.auto = general.OptionalDefaultedBool(true, b...)
	return w
}

func (w *Block) SetPayload(p any) *Block {
	w.payload = p
	return w
}

func (w *Block) Payload() any {
	return w.payload
}

func (w *Block) Reset() {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	w.buf.Reset()
}

// Write save the contents of buf to the writer b. The only errors returned are ones encountered while writing to the underlying buffer.
func (w *Block) Write(buf []byte) (n int, err error) {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	if w.closed {
		return 0, os.ErrClosed
	}
	n, err = w.buf.Write(buf)
	if w.auto {
		w.blocks.requestFlush()
	}
	return n, err
}

func (w *Block) Close() error {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	if w.closed {
		return os.ErrClosed
	}
	w.closed = true
	return w.blocks.discardBlock()
}

func (w *Block) IsClosed() bool {
	w.blocks.lock.Lock()
	defer w.blocks.lock.Unlock()
	return w.closed
}

func (w *Block) Flush() error {
	return w.blocks.Flush()
}

func (w *Block) flush(final bool) (int, error) {
	lines := 0
	titleline := 0
	newline := false
	data := w.buf.Bytes()
	if w.closed && w.final != nil {
		data = w.final
	} else {
		if w.titleline != "" {
			w.blocks.out.Write([]byte(w.titleline + "\n"))
			titleline = 1
		}
	}

	linestart := make([]int, w.view)

	var col int
	start := 0
	for o, b := range data {
		if b == '\n' {
			linestart[lines%w.view] = start
			start = o + 1
			lines++
			newline = true
			col = 0
		} else {
			newline = false
			col++
			if w.blocks.overFlowHandled && col > w.blocks.termWidth {
				lines++
				col = 0
			}
		}
	}

	if !newline {
		linestart[lines%w.view] = start
		lines++
		data = append(data, '\n')
	}

	var err error
	if final || lines <= w.view {
		_, err = w.blocks.out.Write(data)
		return lines + titleline, err
	} else {
		index := (lines) % w.view
		start := linestart[index]
		view := data[start:]
		_, err = w.blocks.out.Write(view)
		return w.view + titleline, err
	}
}
