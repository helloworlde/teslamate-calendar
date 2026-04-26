package calendar

import (
	"fmt"
	"log"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
)

func UpdateEvents(carID, vehicleName string, updates []model.Update, detail bool, dashboardBaseURL, dashboardPath string) []Event {
	out := make([]Event, 0, len(updates))
	for _, u := range updates {
		if u.StartDate == nil {
			log.Printf("debug: skip update without time: id=%v version=%s", u.ID, u.Version)
			continue
		}
		start := u.StartDate.UTC()
		end := start.Add(time.Minute)
		if u.EndDate != nil {
			end = u.EndDate.UTC()
		}
		summary := "⬆️ " + vehicleName + " · 软件更新"
		if u.Version != "" {
			summary += " " + u.Version
		}
		descParts := []string{}
		if u.ID != nil {
			descParts = append(descParts, fmt.Sprintf("更新ID: %v", u.ID))
		}
		descParts = append(descParts, "开始: "+start.Format(time.RFC3339))
		if !end.Equal(start) {
			descParts = append(descParts, "完成: "+end.Format(time.RFC3339))
			descParts = append(descParts, "耗时: "+formatSeconds(end.Sub(start).Seconds()))
		}
		if u.Status != "" {
			descParts = append(descParts, "状态: "+u.Status)
		}
		if detail && u.ReleaseNotes != "" {
			descParts = append(descParts, "Release Notes: "+u.ReleaseNotes)
		}
		desc := AppendLinksSection(strings.Join(descParts, "\n"), "", DashboardURL(dashboardBaseURL, dashboardPath, carID, start, end))
		uid := UIDWithFallback(
			fmt.Sprintf("teslamate-calendar-car-%s-update-", carID),
			u.ID,
			carID, u.Version, start.Format(time.RFC3339),
		)
		out = append(out, Event{
			UID:         uid,
			Start:       start,
			End:         end,
			Summary:     summary,
			Description: desc,
			Location:    "Tesla",
		})
	}
	return out
}
