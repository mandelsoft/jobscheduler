package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	bars := []int{1000, 1004, 1003, 1002}

	for i, b := range bars {
		bar := uiprogress.NewSpinner(p, b).
			//	SetFinal(fmt.Sprintf("action %d done", i+1)).
			SetSpeed(1).
			PrependFunc(uiprogress.Message(fmt.Sprintf("working on %d ...", i+1))).
			AppendElapsed()

		go func() {
			time.Sleep(time.Second * time.Duration(10+rand.Int()%20))
			bar.Close()
		}()
	}

	p.Close()
	p.Wait(nil)
}
