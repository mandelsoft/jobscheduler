package uiprogress

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
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
