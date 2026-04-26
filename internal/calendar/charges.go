package calendar

import (
	"fmt"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
)

func ChargeEvents(carID, vehicleName string, charges []model.Charge, view string, detail bool, dashboardBaseURL, dashboardPath string) []Event {
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
		desc := buildChargeDescription(c, detail)
		location := chargeLocation(c)
		mapURL := GooglePointURL(c.Lat, c.Lng)
		if mapURL == "" {
			mapURL = GoogleSearchURL(location)
		}
		desc = AppendLinksSection(desc, mapURL, DashboardURL(dashboardBaseURL, dashboardPath, carID, start, end))
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
		if view == "detail" {
			kwh = fmt.Sprintf("+%.1fkWh", *c.KwhAdded)
		} else {
			kwh = fmt.Sprintf("+%.0fkWh", *c.KwhAdded)
		}
	}
	batt := ""
	if c.StartBatteryLevel != nil && c.EndBatteryLevel != nil {
		if view == "normal" {
			batt = fmt.Sprintf("%.0f→%.0f%%", *c.StartBatteryLevel, *c.EndBatteryLevel)
		} else {
			batt = fmt.Sprintf("%.0f%%→%.0f%%", *c.StartBatteryLevel, *c.EndBatteryLevel)
		}
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

func buildChargeDescription(c model.Charge, detail bool) string {
	parts := []string{}
	if c.ID != nil {
		parts = append(parts, fmt.Sprintf("充电ID: %v", c.ID))
	}
	if c.StartDate != nil {
		parts = append(parts, "开始: "+c.StartDate.UTC().Format(time.RFC3339))
	}
	if c.EndDate != nil {
		parts = append(parts, "结束: "+c.EndDate.UTC().Format(time.RFC3339))
	}
	if dur := chargeDurationText(c, time.Time{}, time.Time{}); dur != "" {
		parts = append(parts, "耗时: "+dur)
	}
	if c.StartDate != nil && c.EndDate != nil {
		parts = append(parts, fmt.Sprintf("时段: %s -> %s", c.StartDate.In(time.Local).Format("01-02 15:04"), c.EndDate.In(time.Local).Format("01-02 15:04")))
	}
	if !detail {
		return strings.Join(parts, "\n")
	}
	if c.KwhAdded != nil {
		parts = append(parts, fmt.Sprintf("电池增量: %.2f kWh", *c.KwhAdded))
	}
	if wall := rawFloat(c.Raw, "charge_energy_used"); wall != nil {
		parts = append(parts, fmt.Sprintf("墙端电量: %.2f kWh", *wall))
		if c.KwhAdded != nil && *wall > 0 {
			parts = append(parts, fmt.Sprintf("充电效率: %.1f%%", (*c.KwhAdded/(*wall))*100))
		}
	}
	if c.Cost != nil {
		parts = append(parts, fmt.Sprintf("费用: %.2f", *c.Cost))
	}
	if c.MaxPower != nil {
		parts = append(parts, fmt.Sprintf("最高功率: %.1f kW", *c.MaxPower))
	}
	if c.AvgPower != nil {
		parts = append(parts, fmt.Sprintf("平均功率: %.1f kW", *c.AvgPower))
	}
	if c.StartBatteryLevel != nil {
		parts = append(parts, fmt.Sprintf("起始电量: %.1f%%", *c.StartBatteryLevel))
	}
	if c.EndBatteryLevel != nil {
		parts = append(parts, fmt.Sprintf("结束电量: %.1f%%", *c.EndBatteryLevel))
	}
	if delta := percentDelta(c.StartBatteryLevel, c.EndBatteryLevel); delta != "" {
		parts = append(parts, "电量变化: "+delta)
	}
	if outside := rawFloat(c.Raw, "outside_temp_avg"); outside != nil {
		parts = append(parts, fmt.Sprintf("外部温度: %.1f°C", *outside))
	}
	if dataPoints := rawFloat(c.Raw, "data_points"); dataPoints != nil {
		parts = append(parts, fmt.Sprintf("数据点: %.0f", *dataPoints))
	}
	if c.AvgPower == nil && c.KwhAdded != nil {
		if dur := chargeDurationHours(c); dur > 0 {
			parts = append(parts, fmt.Sprintf("平均功率(估算): %.1f kW", round1(*c.KwhAdded/dur)))
		}
	}
	if c.Address != "" {
		parts = append(parts, "地点: "+c.Address)
	} else if c.Geofence != "" {
		parts = append(parts, "地点: "+c.Geofence)
	}
	if c.Lat != nil && c.Lng != nil {
		parts = append(parts, fmt.Sprintf("经纬度: %.6f, %.6f", *c.Lat, *c.Lng))
	}
	return strings.Join(parts, "\n")
}

func chargeDurationText(c model.Charge, start, end time.Time) string {
	if c.DurationMinutes != nil && *c.DurationMinutes > 0 {
		return fmt.Sprintf("%.0f 分", *c.DurationMinutes)
	}
	if c.StartDate != nil && c.EndDate != nil {
		sec := c.EndDate.Sub(*c.StartDate).Seconds()
		if sec > 0 {
			return formatSeconds(sec)
		}
	}
	if !start.IsZero() && !end.IsZero() {
		sec := end.Sub(start).Seconds()
		if sec > 0 {
			return formatSeconds(sec)
		}
	}
	return ""
}

func chargeDurationHours(c model.Charge) float64 {
	if c.DurationMinutes != nil && *c.DurationMinutes > 0 {
		return *c.DurationMinutes / 60
	}
	if c.StartDate != nil && c.EndDate != nil {
		h := c.EndDate.Sub(*c.StartDate).Hours()
		if h > 0 {
			return h
		}
	}
	return 0
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
