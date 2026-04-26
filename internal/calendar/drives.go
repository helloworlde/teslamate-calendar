package calendar

import (
	"fmt"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
)

func DriveEvents(carID, vehicleName string, drives []model.Drive, view string, detail bool, dashboardBaseURL, dashboardPath string) []Event {
	out := make([]Event, 0, len(drives))
	for _, d := range drives {
		if d.StartDate == nil {
			continue
		}
		start := d.StartDate.UTC()
		end := start.Add(time.Minute)
		if d.EndDate != nil {
			end = d.EndDate.UTC()
		}
		from := drivePointName(d, true)
		to := drivePointName(d, false)
		summary := buildDriveSummary(d, vehicleName, from, to, start, end, view)
		location := fmt.Sprintf("%s → %s", from, to)
		mapURL := GoogleRouteURL(d.StartLat, d.StartLng, d.EndLat, d.EndLng)
		if mapURL == "" {
			mapURL = GoogleSearchURL(location)
		}
		desc := buildDriveDescription(d, detail)
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
		dist = fmt.Sprintf("%.0fkm", *d.Distance)
	}
	dur := minutesText(d, start, end)
	cons := ""
	if d.Consumption != nil {
		cons = fmt.Sprintf("%.0fWh/km", *d.Consumption)
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

func buildDriveDescription(d model.Drive, detail bool) string {
	parts := []string{}
	parts = append(parts, fmt.Sprintf("路线: %s -> %s", drivePointName(d, true), drivePointName(d, false)))
	if d.ID != nil {
		parts = append(parts, fmt.Sprintf("行程ID: %v", d.ID))
	}
	if d.StartDate != nil {
		parts = append(parts, "开始: "+d.StartDate.UTC().Format(time.RFC3339))
	}
	if d.EndDate != nil {
		parts = append(parts, "结束: "+d.EndDate.UTC().Format(time.RFC3339))
	}
	if d.DurationSeconds != nil {
		parts = append(parts, "耗时: "+formatSeconds(*d.DurationSeconds))
	}
	if d.StartDate != nil && d.EndDate != nil {
		parts = append(parts, fmt.Sprintf("时段: %s -> %s", d.StartDate.In(time.Local).Format("01-02 15:04"), d.EndDate.In(time.Local).Format("15:04")))
	}
	if !detail {
		return strings.Join(parts, "\n")
	}
	if d.Distance != nil {
		parts = append(parts, fmt.Sprintf("距离: %.2f km", *d.Distance))
	}
	if d.AvgSpeed != nil {
		parts = append(parts, fmt.Sprintf("平均速度: %.1f km/h", *d.AvgSpeed))
	}
	if d.MaxSpeed != nil {
		parts = append(parts, fmt.Sprintf("最高速度: %.1f km/h", *d.MaxSpeed))
	}
	if d.AvgSpeed != nil && d.MaxSpeed != nil {
		parts = append(parts, fmt.Sprintf("速度: Avg %.1f / Max %.1f km/h", *d.AvgSpeed, *d.MaxSpeed))
	}
	if d.AvgPower != nil {
		parts = append(parts, fmt.Sprintf("平均功率: %.1f kW", *d.AvgPower))
	}
	if d.MaxPower != nil {
		parts = append(parts, fmt.Sprintf("最大功率: %.1f kW", *d.MaxPower))
	}
	if minPower := rawFloat(d.Raw, "power_min"); minPower != nil && d.MaxPower != nil {
		parts = append(parts, fmt.Sprintf("功率区间: %.1f ~ %.1f kW", *minPower, *d.MaxPower))
	}
	if d.Consumption != nil {
		parts = append(parts, fmt.Sprintf("平均电耗: %.1f Wh/km", *d.Consumption))
	}
	if net := rawFloat(d.Raw, "energy_consumed_net"); net != nil {
		parts = append(parts, fmt.Sprintf("净耗电: %.3f kWh", *net))
	}
	if d.StartBatteryLevel != nil {
		parts = append(parts, fmt.Sprintf("起始电量: %.1f%%", *d.StartBatteryLevel))
	}
	if d.EndBatteryLevel != nil {
		parts = append(parts, fmt.Sprintf("结束电量: %.1f%%", *d.EndBatteryLevel))
	}
	if delta := percentDelta(d.StartBatteryLevel, d.EndBatteryLevel); delta != "" {
		parts = append(parts, "电量变化: "+delta)
	}
	if outside := rawFloat(d.Raw, "outside_temp_avg"); outside != nil {
		parts = append(parts, fmt.Sprintf("外部温度: %.1f°C", *outside))
	}
	if dataPoints := rawFloat(d.Raw, "data_points"); dataPoints != nil {
		parts = append(parts, fmt.Sprintf("数据点: %.0f", *dataPoints))
	}
	if d.StartAddress != "" {
		parts = append(parts, "起点: "+d.StartAddress)
	}
	if d.EndAddress != "" {
		parts = append(parts, "终点: "+d.EndAddress)
	}
	if d.StartLat != nil && d.StartLng != nil {
		parts = append(parts, fmt.Sprintf("起点经纬度: %.6f, %.6f", *d.StartLat, *d.StartLng))
	}
	if d.EndLat != nil && d.EndLng != nil {
		parts = append(parts, fmt.Sprintf("终点经纬度: %.6f, %.6f", *d.EndLat, *d.EndLng))
	}
	return strings.Join(parts, "\n")
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
		return formatSeconds(*d.DurationSeconds)
	}
	sec := end.Sub(start).Seconds()
	if sec > 0 {
		return formatSeconds(sec)
	}
	return ""
}

func formatSeconds(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0f 秒", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%.0f 分", minutes)
	}
	h := int(minutes) / 60
	m := int(minutes) % 60
	return fmt.Sprintf("%d 小时 %d 分", h, m)
}

func minutesText(d model.Drive, start, end time.Time) string {
	sec := end.Sub(start).Seconds()
	if d.DurationSeconds != nil && *d.DurationSeconds > 0 {
		sec = *d.DurationSeconds
	}
	if sec <= 0 {
		return ""
	}
	return fmt.Sprintf("%.0fmin", sec/60)
}

func geoValue(lat, lng *float64) string {
	if lat == nil || lng == nil {
		return ""
	}
	return fmt.Sprintf("%.6f;%.6f", *lat, *lng)
}
