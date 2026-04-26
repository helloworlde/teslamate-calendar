package model

import "strings"

func GrafanaBaseURLFromGlobalSettings(gs GlobalSettings) string {
	if gs.Raw == nil {
		return ""
	}
	return grafanaBaseFromMap(gs.Raw)
}

func grafanaBaseFromMap(m map[string]any) string {
	keys := []string{"teslamate_urls", "TeslaMateURLs", "teslamateUrls"}
	for _, k := range keys {
		if v, ok := m[k].(map[string]any); ok {
			if s := firstNonEmptyString(v, "grafana_url", "GrafanaURL", "grafanaUrl"); s != "" {
				return s
			}
		}
	}
	return firstNonEmptyString(m, "grafana_url", "GrafanaURL", "grafanaUrl")
}

func firstNonEmptyString(m map[string]any, fieldNames ...string) string {
	for _, name := range fieldNames {
		if s, ok := m[name].(string); ok {
			if t := strings.TrimSpace(s); t != "" {
				return t
			}
		}
	}
	return ""
}
