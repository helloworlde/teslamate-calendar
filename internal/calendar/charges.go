package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/model"
)

func ChargeEvents(carID, vehicleName string, charges []model.Charge, view string, detail bool, loc *time.Location, dashboardTmpl string) []Event {
	_ = detail
	if loc == nil {
		loc = time.Local
	}
	out := make([]Event, 0, len(charges))
	for _, c := range charges {
		if c.StartDate == nil {
			continue
		}
		start := c.StartDate.UTC()
		end := start.Add(time.Minute)
		if c.EndDate != nil {
			end = c.EndDate.UTC()
		}
		summary := buildChargeSummary(c, vehicleName, start, end, view)
		desc := BuildChargeDescription(vehicleName, c, loc)
		location := chargeLocation(c)
		mapURL := GooglePointURL(c.Lat, c.Lng)
		if mapURL == "" {
			mapURL = GoogleSearchURL(location)
		}
		dash := RenderDashboardURL(dashboardTmpl, timesToDashboardArgs(carID, "charges", eventIDString(c.ID), start, end))
		desc = AppendLinksSection(desc, mapURL, dash)
		uid := UIDWithFallback(
			fmt.Sprintf("teslamate-calendar-car-%s-charge-", carID),
			c.ID,
			carID, fmt.Sprintf("%v", c.ID), start.Format(time.RFC3339), end.Format(time.RFC3339),
		)
		out = append(out, Event{
			UID:         uid,
			Start:       start,
			End:         end,
			Summary:     summary,
			Description: desc,
			Location:    location,
			Geo:         geoValue(c.Lat, c.Lng),
			URL:         mapURL,
		})
	}
	return out
}

func buildChargeSummary(c model.Charge, vehicleName string, start, end time.Time, view string) string {
	kwh := ""
	if c.KwhAdded != nil {
		kwh = FormatKWhTitle(*c.KwhAdded)
	}
	batt := ""
	if c.StartBatteryLevel != nil && c.EndBatteryLevel != nil {
		batt = fmt.Sprintf("%.0f%%→%.0f%%", *c.StartBatteryLevel, *c.EndBatteryLevel)
	}
	dur := chargeDurationText(c, start, end)
	switch view {
	case "compact":
		if kwh != "" {
			return "⚡ " + kwh
		}
		return "⚡ 充电"
	case "detail":
		parts := []string{fmt.Sprintf("⚡ %s 充电", vehicleName)}
		if kwh != "" {
			parts = append(parts, kwh)
		}
		if batt != "" {
			parts = append(parts, batt)
		}
		if dur != "" {
			parts = append(parts, dur)
		}
		return strings.Join(parts, " · ")
	default:
		parts := []string{fmt.Sprintf("⚡ %s", vehicleName)}
		if kwh != "" {
			parts = append(parts, kwh)
		}
		if batt != "" {
			parts = append(parts, batt)
		}
		return strings.Join(parts, " · ")
	}
}

func chargeDurationText(c model.Charge, start, end time.Time) string {
	if c.DurationMinutes != nil && *c.DurationMinutes > 0 {
		return FormatDurationZH(*c.DurationMinutes * 60)
	}
	if c.StartDate != nil && c.EndDate != nil {
		sec := c.EndDate.Sub(*c.StartDate).Seconds()
		if sec > 0 {
			return FormatDurationZH(sec)
		}
	}
	if !start.IsZero() && !end.IsZero() {
		sec := end.Sub(start).Seconds()
		if sec > 0 {
			return FormatDurationZH(sec)
		}
	}
	return ""
}

func chargeLocation(c model.Charge) string {
	if strings.TrimSpace(c.Address) != "" {
		return c.Address
	}
	if strings.TrimSpace(c.Geofence) != "" {
		return c.Geofence
	}
	return "Charging"
}
