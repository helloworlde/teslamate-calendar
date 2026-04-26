package util

import "fmt"

func F(v *float64, format string) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf(format, *v)
}
