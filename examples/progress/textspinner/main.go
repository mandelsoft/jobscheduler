package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	for s := 0; s < 3; s++ {
		bar := uiprogress.NewTextSpinner(p, 5, 3).
			PrependFunc(uiprogress.Message(fmt.Sprintf("working on task %d...", s+1))).
			AppendElapsed()

		go func() {
			// starts automatically, with the first write
			steps := 6 + rand.Int()%10
			for i := 0; i <= steps; i++ {
				t := 500 + 200*(rand.Int()%6)
				fmt.Fprintf(bar, "doing step %d[%dms]\n", i, t)
				time.Sleep(time.Duration(t) * time.Millisecond)
			}
			bar.Close()
		}()
	}
	p.Close()
	p.Wait(nil)
}
