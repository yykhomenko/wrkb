package wrkb

import (
	"fmt"
	"time"
)

func formatDuration1(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond))
	case d >= time.Microsecond:
		return fmt.Sprintf("%.1fµs", float64(d)/float64(time.Microsecond))
	default:
		return fmt.Sprintf("%.1fns", float64(d))
	}
}
