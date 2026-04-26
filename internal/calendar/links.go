package calendar

import (
	"fmt"
	"strings"
	"time"
)

func DashboardURL(baseURL, path, carID string, start, end time.Time) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	if end.IsZero() || !end.After(start) {
		end = start.Add(time.Hour)
	}
	if strings.TrimSpace(path) == "" {
		return baseURL
	}
	return fmt.Sprintf("%s%s?from=%d&to=%d&var-car_id=%s", baseURL, path, start.UnixMilli(), end.UnixMilli(), carID)
}

func AppendLinksSection(desc, mapURL, dashboardURL string) string {
	links := []string{}
	if mapURL != "" {
		links = append(links, "地图："+mapURL)
	}
	if dashboardURL != "" {
		links = append(links, "TeslaMate 看板："+dashboardURL)
	}
	if len(links) == 0 {
		return desc
	}
	parts := []string{desc, "", "━━━━━━━━━━━━", "🔗 链接", "━━━━━━━━━━━━"}
	parts = append(parts, links...)
	return strings.Join(parts, "\n")
}
