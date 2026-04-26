package calendar

import "strings"

func EscapeText(v string) string {
	r := strings.ReplaceAll(v, "\\", "\\\\")
	r = strings.ReplaceAll(r, ";", "\\;")
	r = strings.ReplaceAll(r, ",", "\\,")
	r = strings.ReplaceAll(r, "\r\n", "\n")
	r = strings.ReplaceAll(r, "\r", "\n")
	r = strings.ReplaceAll(r, "\n", "\\n")
	return r
}
