package main

import (
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/ttyprogress"
)

func main() {
	p := ttyprogress.New(os.Stdout)

	bar, _ := ttyprogress.NewBar().
		SetPredefined(10).
		SetWidth(ttyprogress.PercentTerminalSize(30)).
		PrependFunc(ttyprogress.Message("Downloading...")).PrependElapsed().AppendCompleted().
		Add(p)

	for i := 0; i <= 20; i++ {
		bar.Set(i * 5)
		time.Sleep(time.Millisecond * 500)
	}
}
