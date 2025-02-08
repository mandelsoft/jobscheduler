package main

import (
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	bar := uiprogress.NewSteps(p, "downloading", "unpacking", "installing", "verifying").PrependFunc(uiprogress.Message("progressbar"), 0).PrependElapsed().AppendCompleted()

	bar.Start()
	for i := 0; i < 4; i++ {
		time.Sleep(time.Second * 2)
		bar.Incr()
	}
}
