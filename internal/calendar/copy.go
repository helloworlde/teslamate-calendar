package calendar

import "strings"

func Section(title string, lines ...string) []string {
	out := []string{
		"",
		"━━━━━━━━━━━━",
		title,
		"━━━━━━━━━━━━",
	}
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		out = append(out, ln)
	}
	return out
}
