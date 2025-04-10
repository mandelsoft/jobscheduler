package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/ttycolors"
	"github.com/mandelsoft/ttyprogress"
)

func main() {
	p := ttyprogress.For(os.Stdout)

	bars := []int{1000, 1002, 1003}
	cols := []ttycolors.Format{
		ttycolors.New(ttycolors.FmtBrightGreen, ttycolors.FmtUnderline),
		ttycolors.New(ttycolors.FmtCyan, ttycolors.FmtItalic),
		ttycolors.New(ttycolors.FmtBgCyan, ttycolors.FmtBold),
	}
	for i, b := range bars {
		bar, _ := ttyprogress.NewSpinner().
			SetPredefined(b).
			SetSpeed(1).
			SetColor(cols[i]).
			PrependFunc(ttyprogress.Message(fmt.Sprintf("working on task %d ...", i+1))).
			AppendElapsed().Add(p)
		bar.Start()
		go func() {
			time.Sleep(time.Second * time.Duration(10+rand.Int()%20))
			bar.Close()
		}()
	}

	p.Close()
	p.Wait(nil)
}

func runner(id string, e ttyprogress.Element, txt io.WriteCloser) {
	e.Start()
	lines := rand.Intn(10) + 10
	for i := 0; i < lines; i++ {
		time.Sleep(time.Duration((500 + rand.Intn(2500))) * time.Millisecond)
		fmt.Fprintf(txt, "job %s line %d\n", id, i+1)
	}
	txt.Close()
	e.Close()
}

type dummy struct{}

func (d dummy) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (d dummy) Close() error {
	return nil
}
