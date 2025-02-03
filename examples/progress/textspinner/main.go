package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	bar := uiprogress.NewTextSpinner(p, 5, 3).
		PrependFunc(uiprogress.Message(fmt.Sprintf("working on task ..."))).
		AppendElapsed()

	for i := 0; i <= 20; i++ {
		fmt.Fprintf(bar, "doing step %d\n", i)
		time.Sleep(time.Millisecond * 500)
	}
	bar.Close()
}
