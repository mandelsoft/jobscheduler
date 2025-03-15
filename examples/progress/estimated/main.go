package main

import (
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	total := 10 * time.Second
	bar := uiprogress.NewEstimated(p, total).SetPredefined(10).
		PrependFunc(uiprogress.Message("Downloading...")).PrependEstimated().
		AppendCompleted().AppendElapsed().SetWidth(uiprogress.ReserveTerminalSize(40))
	bar.Start()
	p.Close()

	for i := 0; i <= 19; i++ {
		time.Sleep(time.Millisecond * 500)
		// Adjust expected duration
		total = total + 100*time.Millisecond
		bar.Set(total)
	}
	time.Sleep(time.Second * 2)
	bar.Close()
}
