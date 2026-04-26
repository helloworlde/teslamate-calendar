package calendar

import (
	"fmt"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
)

func DriveEvents(carID, vehicleName string, drives []model.Drive, view string, detail bool, loc *time.Location, dashboardBaseURL, dashboardPath string) []Event {
	_ = detail
	out := make([]Event, 0, len(drives))
	if loc == nil {
		loc = time.Local
	}
	for _, d := range drives {
		if d.StartDate == nil {
			continue
		}
		start := d.StartDate.UTC()
		end := start.Add(time.Minute)
		if d.EndDate != nil {
			end = d.EndDate.UTC()
		}
		from := NormalizePlaceName(drivePointName(d, true))
		to := NormalizePlaceName(drivePointName(d, false))
		summary := buildDriveSummary(d, vehicleName, from, to, start, end, view)
		location := fmt.Sprintf("%s → %s", from, to)
		mapURL := GoogleRouteURL(d.StartLat, d.StartLng, d.EndLat, d.EndLng)
		if mapURL == "" {
			mapURL = GoogleSearchURL(location)
		}
		desc := BuildDriveDescription(vehicleName, d, loc)
		desc = AppendLinksSection(desc, mapURL, DashboardURL(dashboardBaseURL, dashboardPath, carID, start, end))
		uid := UIDWithFallback(
			fmt.Sprintf("teslamate-calendar-car-%s-drive-", carID),
			d.ID,
			carID, fmt.Sprintf("%v", d.ID), start.Format(time.RFC3339), end.Format(time.RFC3339),
		)
		out = append(out, Event{
			UID:         uid,
			Start:       start,
			End:         end,
			Summary:     summary,
			Description: desc,
			Location:    location,
			Geo:         geoValue(d.StartLat, d.StartLng),
			URL:         mapURL,
		})
	}
	return out
}

func buildDriveSummary(d model.Drive, vehicleName, from, to string, start, end time.Time, view string) string {
	dist := ""
	if d.Distance != nil {
		dist = FormatDistanceTitleOneDecimal(*d.Distance)
	}
	dur := ""
	if d.DurationSeconds != nil {
		dur = FormatDurationZH(*d.DurationSeconds)
	} else if d.StartDate != nil && d.EndDate != nil {
		dur = FormatDurationZH(d.EndDate.Sub(*d.StartDate).Seconds())
	} else {
		dur = FormatDurationZH(end.Sub(start).Seconds())
	}
	cons := ""
	if d.Consumption != nil {
		cons = FormatWhPerKm(*d.Consumption)
	}
	switch view {
	case "compact":
		parts := []string{"🚗"}
		if dist != "" {
			parts = append(parts, dist)
		}
		if dur != "" {
			parts = append(parts, dur)
		}
		return strings.Join(parts, " · ")
	case "detail":
		parts := []string{fmt.Sprintf("🚗 %s · %s → %s", vehicleName, from, to)}
		if dist != "" {
			parts = append(parts, dist)
		}
		if dur != "" {
			parts = append(parts, dur)
		}
		if cons != "" {
			parts = append(parts, cons)
		}
		return strings.Join(parts, " · ")
	default:
		parts := []string{fmt.Sprintf("🚗 %s · %s → %s", vehicleName, from, to)}
		if dist != "" {
			parts = append(parts, dist)
		}
		return strings.Join(parts, " · ")
	}
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func drivePointName(d model.Drive, isStart bool) string {
	if isStart {
		if strings.TrimSpace(d.StartAddress) != "" {
			return d.StartAddress
		}
		if strings.TrimSpace(d.StartGeofence) != "" {
			return d.StartGeofence
		}
		return "起点"
	}
	if strings.TrimSpace(d.EndAddress) != "" {
		return d.EndAddress
	}
	if strings.TrimSpace(d.EndGeofence) != "" {
		return d.EndGeofence
	}
	return "终点"
}

func driveDurationText(d model.Drive, start, end time.Time) string {
	if d.DurationSeconds != nil && *d.DurationSeconds > 0 {
		return FormatDurationZH(*d.DurationSeconds)
	}
	sec := end.Sub(start).Seconds()
	if sec > 0 {
		return FormatDurationZH(sec)
	}
	return ""
}

func minutesText(d model.Drive, start, end time.Time) string {
	sec := end.Sub(start).Seconds()
	if d.DurationSeconds != nil && *d.DurationSeconds > 0 {
		sec = *d.DurationSeconds
	}
	if sec <= 0 {
		return ""
	}
	return FormatDurationZH(sec)
}

func geoValue(lat, lng *float64) string {
	if lat == nil || lng == nil {
		return ""
	}
	return fmt.Sprintf("%.6f;%.6f", *lat, *lng)
}
