package model

import "strings"

func GrafanaBaseURLFromGlobalSettings(gs GlobalSettings) string {
	if s := strings.TrimSpace(gs.GrafanaURL); s != "" {
		return s
	}
	return ""
}
