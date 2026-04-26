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

func JoinNonEmpty(sep string, parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(sep)
		}
		b.WriteString(p)
	}
	return b.String()
}
