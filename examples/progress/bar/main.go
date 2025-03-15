package main

import (
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	bar := uiprogress.NewBar(p, 100).SetPredefined(10).SetWidth(uiprogress.PercentTerminalSize(30)).
		PrependFunc(uiprogress.Message("Downloading...")).PrependElapsed().AppendCompleted()

	for i := 0; i <= 20; i++ {
		bar.Set(i * 5)
		time.Sleep(time.Millisecond * 500)
	}
}
