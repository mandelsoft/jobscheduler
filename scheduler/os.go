package scheduler

import (
	"io"
	"os"
	"sync"
)

var Stdout = &SyncedWriter{Writer: os.Stdout}
var Stderr = &SyncedWriter{Writer: os.Stderr}

type SyncedWriter struct {
	lock sync.Mutex
	io.Writer
}

func (w *SyncedWriter) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.Writer.Write(p)
}
