package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
	"github.com/mandelsoft/jobscheduler/units"
)

func main() {
	p := uiprogress.New(os.Stdout)

	for b := 0; b < 3; b++ {
		total := 100 + rand.Int()%100
		w := uiprogress.NewBar(p, total).
			SetPredefined(1).
			SetFinal(fmt.Sprintf("Finished: Downloaded %d GB", total)).
			AppendFunc(uiprogress.Amount(units.Plain)).
			AppendFunc(uiprogress.Message("GB")).
			PrependFunc(uiprogress.Message("Downloading ..."))

		go func() {
			done := 0
			for w.Set(done) {
				time.Sleep(time.Millisecond * time.Duration(100+rand.Int()%500))
				done = done + 5
			}
		}()
	}
	p.Close()
	p.Wait(nil)
}
