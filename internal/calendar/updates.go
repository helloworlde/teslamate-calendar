package calendar

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/model"
)

func UpdateEvents(carID, vehicleName string, updates []model.Update, detail bool, loc *time.Location, dashboardTmpl string) []Event {
	if loc == nil {
		loc = time.Local
	}
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
		if u.Version != "" {
			descParts = append(descParts, "版本："+u.Version)
		}
		st := u.StartDate.In(loc)
		if u.EndDate != nil {
			et := u.EndDate.In(loc)
			descParts = append(descParts, "时间："+FormatTimeRange(*u.StartDate, *u.EndDate, loc))
			if d := et.Sub(st).Seconds(); d > 0 {
				if z := FormatDurationZH(d); z != "" {
					descParts = append(descParts, "耗时："+z)
				}
			}
		} else {
			descParts = append(descParts, "时间："+st.Format("2006-01-02 15:04"))
		}
		if u.Status != "" {
			descParts = append(descParts, "状态："+u.Status)
		}
		if detail && u.ReleaseNotes != "" {
			descParts = append(descParts, "Release Notes: "+u.ReleaseNotes)
		}
		dash := RenderDashboardURL(dashboardTmpl, timesToDashboardArgs(carID, "updates", eventIDString(u.ID), start, end))
		desc := AppendLinksSection(strings.Join(descParts, "\n"), "", dash)
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
