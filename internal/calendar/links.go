package calendar

import (
	"fmt"
	"strings"
	"time"
)

type DashboardArgs struct {
	FromMs  int64
	ToMs    int64
	CarID   string
	Range   string
	EventID string
}

func eventIDString(id any) string {
	if id == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", id))
	if s == "" || s == "<nil>" {
		return ""
	}
	return s
}

func RenderDashboardURL(template string, a DashboardArgs) string {
	t := strings.TrimSpace(template)
	if t == "" {
		return ""
	}
	to := a.ToMs
	if to <= a.FromMs {
		to = a.FromMs + 60*60*1000
	}
	rep := strings.NewReplacer(
		"{from}", fmt.Sprintf("%d", a.FromMs),
		"{to}", fmt.Sprintf("%d", to),
		"{car_id}", a.CarID,
		"{range}", a.Range,
		"{event_id}", a.EventID,
	)
	return rep.Replace(t)
}

func timesToDashboardArgs(carID, rangeKey, eventID string, start, end time.Time) DashboardArgs {
	fromMs := start.UnixMilli()
	toMs := end.UnixMilli()
	if toMs <= fromMs {
		toMs = fromMs + 60*1000
	}
	return DashboardArgs{
		FromMs:  fromMs,
		ToMs:    toMs,
		CarID:   carID,
		Range:   rangeKey,
		EventID: eventID,
	}
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
