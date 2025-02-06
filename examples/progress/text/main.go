package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiprogress"
)

func main() {
	p := uiprogress.New(os.Stdout)

	text := uiprogress.NewText(p, 3).SetAuto()

	for i := 0; i <= 20; i++ {
		fmt.Fprintf(text, "doing step %d\n", i)
		time.Sleep(time.Millisecond * 500)
	}
	text.Close()
}
