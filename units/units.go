package units

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
)

type Unit = func(n int) string

func Plain(n int) string {
	return fmt.Sprintf("%d", n)
}

func Scaled(v int, factor int, units []string, scale ...int64) string {
	var s int64
	n := int64(v) * general.OptionalDefaulted[int64](int64(1), scale...)
	for _, u := range units {
		n, s = n/int64(factor), n
		if n == 0 {
			if u == "" {
				return fmt.Sprintf("%d", s)

			} else {
				return fmt.Sprintf("%d %s", s, u)
			}
		}

	}
	u := units[len(units)-1]
	if u == "" {
		return fmt.Sprintf("%d", s)
	} else {
		return fmt.Sprintf("%d %s", s, u)
	}
}

var byteUnits = []string{"", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

func Bytes(scale ...int64) Unit {
	return func(n int) string {
		return Scaled(n, 1024, byteUnits, scale...)
	}
}

var lengthUnits = []string{"mm", "m", "km"}

func Millimeter(scale ...int64) Unit {
	return func(n int) string {
		return Scaled(n, 1000, lengthUnits, scale...)
	}
}

func Seconds(n int) string {
	m, s := n/60, n%60
	if m == 0 {
		return fmt.Sprintf("%ds", s)
	}

	h, m := m/60, m%60
	if h == 0 {
		return fmt.Sprintf("%d:%d", m, s)
	}

	d, h := h/24, h%24
	if d == 0 {
		return fmt.Sprintf("%d:%d:%d", h, m, s)
	}
	return fmt.Sprintf("%d days %d:%d:%d", d, h, m, s)
}
