package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mandelsoft/jobscheduler/uiblocks"
)

func main() {
	blocks := uiblocks.New(os.Stdout)

	writer := blocks.NewBlock(3).SetAuto().SetTitleLine("Some work:")

	for i := 0; i <= 20; i++ {
		fmt.Fprintf(writer, "doing step %d\n", i)
		time.Sleep(time.Millisecond * 500)
	}

	writer.Close()
}
