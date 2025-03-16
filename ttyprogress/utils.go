package ttyprogress

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/jobscheduler/uiblocks"
	"github.com/mandelsoft/jobscheduler/units"
)

////////////////////////////////////////////////////////////////////////////////

func Message(m string) DecoratorFunc {
	return func(element Element) string {
		return m
	}
}

func Amount(unit ...units.Unit) func(Element) string {
	u := general.OptionalDefaulted(units.Plain, unit...)
	return func(e Element) string {
		if t, ok := e.(interface{ Total() int }); ok {
			return fmt.Sprintf("(%s/%s)", u(e.(Bar).Current()), u(t.Total()))
		}
		return fmt.Sprintf("(%s)", u(e.(Bar).Current()))
	}
}

// PercentTerminalSize return a width relative to to the terminal size.
func PercentTerminalSize(p uint) uint {
	x, _ := uiblocks.GetTerminalSize()

	if x == 0 {
		return 10
	}
	s := (uint(x) * p) / 100
	if s < 10 {
		return 10
	}
	return s
}

// ReserveTerminalSize provide a reasonable width
// reserving an amount of characters for predefined fixed
// content.
func ReserveTerminalSize(r uint) uint {
	x, _ := uiblocks.GetTerminalSize()
	if x == 0 {
		return 10
	}
	s := x - int(r)
	if s < 10 {
		return 10
	}
	return uint(s)
}
