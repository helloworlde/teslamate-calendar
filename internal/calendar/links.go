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
	mapURL = strings.TrimSpace(mapURL)
	dashboardURL = strings.TrimSpace(dashboardURL)
	if mapURL == "" && dashboardURL == "" {
		return desc
	}
	parts := []string{}
	if strings.TrimSpace(desc) != "" {
		parts = append(parts, desc, "")
	}
	parts = append(parts, "━━━━━━━━━━━━", "🔗 相关链接", "━━━━━━━━━━━━")
	if mapURL != "" {
		parts = append(parts, "地图：", mapURL)
	}
	if dashboardURL != "" {
		parts = append(parts, "TeslaMate 看板：", dashboardURL)
	}
	return strings.Join(parts, "\n")
}
