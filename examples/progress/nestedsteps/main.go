package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func Step(n string) uiprogress.NestedStep {
	return uiprogress.NestedStep{
		Name: n,
		Factory: func(p uiprogress.Container, n string) uiprogress.Element {
			return uiprogress.NewBar(p, 100).
				PrependFunc(uiprogress.Message(n)).
				PrependElapsed().
				AppendCompleted() // .SetFinal("")  // show only current step
		},
	}
}

func main() {
	p := uiprogress.New(os.Stdout)

	bar := uiprogress.NewNestedSteps(p, "  ", false,
		Step("downloading"),
		Step("unpacking"),
		Step("installing"),
		Step("verifying")).
		PrependFunc(uiprogress.Message("progressbar"), 0).PrependElapsed().AppendCompleted()

	go func() {
		e := bar.Start()
		for i := 0; i < 4; i++ {
			for i := 0; i < 100; i++ {
				time.Sleep(time.Millisecond * time.Duration(rand.Int()%100))
				e.(uiprogress.Bar).Incr()
			}
			e = bar.Incr()
		}
	}()

	bar.Wait(nil)
}
