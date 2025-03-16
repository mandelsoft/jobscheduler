package specs

import (
	"fmt"
)

// PercentString returns the formatted string representation of the percent value.
func PercentString(p float64) string {
	return fmt.Sprintf("%3.f%%", p)
}
