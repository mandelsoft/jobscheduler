package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func grouped(p uiprogress.Container, lvl int) {
	g := uiprogress.NewGroup(p, "- ", 86).
		SetFollowUpGap("  ").
		SetSpeed(5).
		PrependFunc(uiprogress.Message(fmt.Sprintf("Grouped work %d", lvl))).
		AppendElapsed()
	if lvl > 0 {
		grouped(g, lvl-1)
	}
	for i := 0; i < 2; i++ {
		bar := uiprogress.NewSpinner(g, 70).
			SetSpeed(1).
			PrependFunc(uiprogress.Message(fmt.Sprintf("working on task %d[%d]...", i+1, lvl))).
			AppendElapsed()

		go func() {
			time.Sleep(time.Second * time.Duration(10+rand.Int()%20))
			bar.Close()
		}()
	}

	text := uiprogress.NewTextSpinner(g, 70, 3).
		SetSpeed(1).
		SetGap("  ").
		PrependFunc(uiprogress.Message(fmt.Sprintf("working on task %d[%d]...", 3, lvl))).
		AppendElapsed()

	go func() {
		for i := 0; i <= 20; i++ {
			fmt.Fprintf(text, "doing step %d of task %d[%d]\n", i, 3, lvl)
			time.Sleep(time.Millisecond * 100 * time.Duration(1+rand.Int()%20))
		}
		text.Close()
	}()
	g.Close()
}

func main() {
	p := uiprogress.New(os.Stdout)
	grouped(p, 2)
	p.Close()
	p.Wait(nil)
}
