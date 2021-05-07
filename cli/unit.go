package cli

import "fmt"

func DisplayUnit(value float64) string {
	if value > 1000*1000*1000 {
		return fmt.Sprintf("%.2fT", value/1000000000)
	}
	if value > 1000*1000 {
		return fmt.Sprintf("%.2fM", value/1000000)
	}
	if value > 1000 {
		return fmt.Sprintf("%.2fk", value/1000)
	}
	return fmt.Sprintf("%.2f", value)
}
