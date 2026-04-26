package calendar

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
)

func driveHasRouteRecorded(d model.Drive) bool {
	if d.Raw == nil {
		return false
	}
	keys := []string{"polyline", "osrm_polyline", "Polyline", "path"}
	for _, k := range keys {
		v, ok := d.Raw[k]
		if !ok {
			continue
		}
		switch s := v.(type) {
		case string:
			if len(strings.TrimSpace(s)) > 8 {
				return true
			}
		}
	}
	if s := rawString(d.Raw, "route", "polyline"); len(strings.TrimSpace(s)) > 8 {
		return true
	}
	return false
}

func BuildDailyDescription(vehicleName string, d model.DailySummary, includeItems, detail bool, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	lines := []string{
		fmt.Sprintf("%s · %s", d.Day.In(loc).Format("2006-01-02"), vehicleName),
	}
	var overview []string
	if d.DriveCount > 0 {
		parts := []string{fmt.Sprintf("🚗 行程 %d 次", d.DriveCount)}
		if s := FormatDistanceDetail(d.Distance); s != "" {
			parts = append(parts, s)
		}
		if s := FormatDurationZH(d.DriveSeconds); s != "" {
			parts = append(parts, s)
		}
		overview = append(overview, strings.Join(parts, " · "))
	}
	if d.ChargeCount > 0 {
		parts := []string{fmt.Sprintf("⚡ 充电 %d 次", d.ChargeCount)}
		if s := FormatKWhDetail(d.KwhAdded); s != "" {
			parts = append(parts, s)
		}
		if s := FormatDurationZH(d.ChargeSeconds); s != "" {
			parts = append(parts, s)
		}
		overview = append(overview, strings.Join(parts, " · "))
	}
	if d.StartBattery != nil && d.EndBattery != nil {
		if br := FormatBatteryRange(d.StartBattery, d.EndBattery, true); br != "" {
			overview = append(overview, "🔋 电量 "+br)
		}
	}
	if d.DriveCount > 0 && d.MaxSpeed > 0 && !isBadFloat(d.MaxSpeed) {
		overview = append(overview, "🏁 最高速度 "+FormatSpeedKMH(d.MaxSpeed))
	}
	if d.Cost > 0 && !isBadFloat(d.Cost) {
		if c := FormatCost(d.Cost); c != "" {
			overview = append(overview, "💰 费用 "+c)
		}
	}
	if len(overview) > 0 {
		lines = append(lines, Section("📌 今日概览", overview...)...)
	}

	if includeItems && d.DriveCount > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "🚗 行程", "━━━━━━━━━━━━")
		for i, block := range d.DriveDetails {
			lines = append(lines, formatNumberedBlock(i+1, block)...)
		}
	}
	if includeItems && d.ChargeCount > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "⚡ 充电", "━━━━━━━━━━━━")
		for i, block := range d.ChargeDetails {
			lines = append(lines, formatNumberedBlock(i+1, block)...)
		}
	}
	if len(d.UpdateVersions) > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "⬆️ 软件更新", "━━━━━━━━━━━━")
		for _, v := range d.UpdateVersions {
			if strings.TrimSpace(v) != "" {
				lines = append(lines, "版本："+v)
			}
		}
	} else if d.UpdateCount > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "⬆️ 软件更新", "━━━━━━━━━━━━")
		lines = append(lines, fmt.Sprintf("次数：%d 次", d.UpdateCount))
	}
	_ = detail
	return strings.Join(lines, "\n")
}

func formatNumberedBlock(n int, block string) []string {
	block = strings.TrimSpace(block)
	if block == "" {
		return nil
	}
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return nil
	}
	var out []string
	out = append(out, fmt.Sprintf("%d. %s", n, lines[0]))
	for _, ln := range lines[1:] {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		out = append(out, "   "+ln)
	}
	return out
}

func formatDailyDriveBlock(d model.Drive, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	if d.StartDate == nil {
		return ""
	}
	end := *d.StartDate
	if d.EndDate != nil {
		end = *d.EndDate
	}
	tRange := FormatTimeRange(*d.StartDate, end, loc)
	parts1 := []string{tRange}
	if d.DurationSeconds != nil {
		if x := FormatDurationZH(*d.DurationSeconds); x != "" {
			parts1 = append(parts1, x)
		}
	} else {
		if x := FormatDurationZH(end.Sub(*d.StartDate).Seconds()); x != "" {
			parts1 = append(parts1, x)
		}
	}
	if d.Distance != nil {
		if x := FormatDistanceDetail(*d.Distance); x != "" {
			parts1 = append(parts1, x)
		}
	}
	line1 := strings.Join(parts1, " · ")
	from := NormalizePlaceName(drivePointName(d, true))
	to := NormalizePlaceName(drivePointName(d, false))
	line2 := from + " → " + to
	var line3 string
	if d.Consumption != nil {
		line3 = FormatWhPerKm(*d.Consumption)
	}
	if d.StartBatteryLevel != nil && d.EndBatteryLevel != nil {
		if s := FormatBatteryRange(d.StartBatteryLevel, d.EndBatteryLevel, false); s != "" {
			if line3 != "" {
				line3 += " · " + s
			} else {
				line3 = s
			}
		}
	}
	return strings.Join(splitsNonEmpty(line1, line2, line3), "\n")
}

func formatDailyChargeBlock(c model.Charge, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	if c.StartDate == nil {
		return ""
	}
	var end time.Time
	if c.EndDate != nil {
		end = *c.EndDate
	} else {
		end = *c.StartDate
	}
	tRange := FormatTimeRange(*c.StartDate, end, loc)
	parts1 := []string{tRange}
	var sec float64
	if c.DurationMinutes != nil {
		sec = *c.DurationMinutes * 60
	} else if c.StartDate != nil && c.EndDate != nil {
		sec = c.EndDate.Sub(*c.StartDate).Seconds()
	}
	if s := FormatDurationZH(sec); s != "" {
		parts1 = append(parts1, s)
	}
	line1 := strings.Join(parts1, " · ")
	place := c.Address
	if strings.TrimSpace(place) == "" {
		place = c.Geofence
	}
	line2 := NormalizePlaceName(strings.TrimSpace(place))
	var line3 string
	if c.KwhAdded != nil {
		line3 = FormatKWhInDetailLine(*c.KwhAdded)
	}
	if c.StartBatteryLevel != nil && c.EndBatteryLevel != nil {
		if s := FormatBatteryRange(c.StartBatteryLevel, c.EndBatteryLevel, false); s != "" {
			if line3 != "" {
				line3 += " · " + s
			} else {
				line3 = s
			}
		}
	}
	if c.Cost != nil && *c.Cost > 0 && !isBadFloat(*c.Cost) {
		if fc := FormatCost(*c.Cost); fc != "" {
			if line3 != "" {
				line3 += " · " + fc
			} else {
				line3 = fc
			}
		}
	}
	return strings.Join(splitsNonEmpty(line1, line2, line3), "\n")
}

func splitsNonEmpty(s ...string) []string {
	var out []string
	for _, x := range s {
		if strings.TrimSpace(x) == "" {
			continue
		}
		out = append(out, x)
	}
	return out
}

func BuildWeeklyDescription(vehicleName string, weekStart, weekEnd time.Time, loc *time.Location, distance float64, driveCnt, chargeCnt, updateCnt int, kwh float64, maxSpeed, cost, driveSeconds float64, days []model.DailySummary, detail bool) string {
	if loc == nil {
		loc = time.Local
	}
	lastDay := weekEnd.AddDate(0, 0, -1)
	header := fmt.Sprintf("%s ~ %s · %s", weekStart.In(loc).Format("2006-01-02"), lastDay.In(loc).Format("2006-01-02"), vehicleName)
	lines := []string{header}
	var overview []string
	if driveCnt > 0 {
		p := []string{fmt.Sprintf("🚗 行程 %d 次", driveCnt)}
		if s := FormatDistanceDetail(distance); s != "" {
			p = append(p, s)
		}
		if s := FormatDurationZH(driveSeconds); s != "" {
			p = append(p, s)
		}
		overview = append(overview, strings.Join(p, " · "))
	}
	if chargeCnt > 0 {
		p := []string{fmt.Sprintf("⚡ 充电 %d 次", chargeCnt)}
		if s := FormatKWhDetail(kwh); s != "" {
			p = append(p, s)
		}
		overview = append(overview, strings.Join(p, " · "))
	}
	if driveCnt > 0 && maxSpeed > 0 {
		overview = append(overview, "🏁 最高速度 "+FormatSpeedKMH(maxSpeed))
	}
	if cost > 0 {
		if c := FormatCost(cost); c != "" {
			overview = append(overview, "💰 费用 "+c)
		}
	}
	lines = append(lines, Section("📌 本周概览", overview...)...)
	_ = updateCnt
	if !detail {
		return strings.Join(lines, "\n")
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Day.Before(days[j].Day)
	})
	var dayLines []string
	for _, d := range days {
		ds := d.Day.In(loc)
		d0 := time.Date(ds.Year(), ds.Month(), ds.Day(), 0, 0, 0, 0, loc)
		w0 := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)
		if d0.Before(w0) || !d0.Before(weekEnd) {
			continue
		}
		p := []string{ds.Format("01-02")}
		if s := FormatDistanceDetail(d.Distance); s != "" {
			p = append(p, s)
		}
		if d.DriveCount > 0 {
			p = append(p, fmt.Sprintf("%d行程", d.DriveCount))
		}
		if d.ChargeCount > 0 {
			p = append(p, fmt.Sprintf("%d充电", d.ChargeCount))
		}
		if d.UpdateCount > 0 {
			p = append(p, fmt.Sprintf("%d更新", d.UpdateCount))
		}
		dayLines = append(dayLines, strings.Join(p, " · "))
	}
	if len(dayLines) > 0 {
		lines = append(lines, Section("📅 每日概览", dayLines...)...)
	}
	return strings.Join(lines, "\n")
}

func BuildMonthlyDescription(vehicleName string, monthStart, monthEnd time.Time, loc *time.Location, distance float64, driveCnt, chargeCnt, updateCnt int, kwh, maxSpeed, cost, driveSeconds float64, days []model.DailySummary, carID string, detail bool) string {
	if loc == nil {
		loc = time.Local
	}
	header := fmt.Sprintf("%s · %s", monthStart.In(loc).Format("2006-01"), vehicleName)
	lines := []string{header}
	var overview []string
	if driveCnt > 0 {
		p := []string{fmt.Sprintf("🚗 行程 %d 次", driveCnt)}
		if s := FormatDistanceDetail(distance); s != "" {
			p = append(p, s)
		}
		if s := FormatDurationZH(driveSeconds); s != "" {
			p = append(p, s)
		}
		overview = append(overview, strings.Join(p, " · "))
	}
	if chargeCnt > 0 {
		p := []string{fmt.Sprintf("⚡ 充电 %d 次", chargeCnt)}
		if s := FormatKWhDetail(kwh); s != "" {
			p = append(p, s)
		}
		overview = append(overview, strings.Join(p, " · "))
	}
	if driveCnt > 0 && maxSpeed > 0 {
		overview = append(overview, "🏁 最高速度 "+FormatSpeedKMH(maxSpeed))
	}
	if cost > 0 {
		if c := FormatCost(cost); c != "" {
			overview = append(overview, "💰 费用 "+c)
		}
	}
	lines = append(lines, Section("📌 本月概览", overview...)...)
	_ = updateCnt
	if !detail {
		return strings.Join(lines, "\n")
	}
	weekLines := monthlyWeekBreakdown(vehicleName, monthStart, monthEnd, loc, carID, days, detail)
	if len(weekLines) > 0 {
		lines = append(lines, Section("📅 每周概览", weekLines...)...)
	}
	return strings.Join(lines, "\n")
}

func monthlyWeekBreakdown(vehicleName string, monthStart, monthEnd time.Time, loc *time.Location, carID string, days []model.DailySummary, detail bool) []string {
	type wk struct {
		start   time.Time
		dist    float64
		drives  int
		charges int
	}
	byKey := map[string]*wk{}
	for _, d := range days {
		ds := d.Day.In(loc)
		if ds.Before(monthStart) || !ds.Before(monthEnd) {
			continue
		}
		ws := startOfWeek(ds)
		if ws.Before(monthStart) {
			ws = monthStart
		}
		key := ws.Format("2006-01-02")
		if _, ok := byKey[key]; !ok {
			byKey[key] = &wk{start: ws}
		}
		w := byKey[key]
		w.dist += d.Distance
		w.drives += d.DriveCount
		w.charges += d.ChargeCount
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil
	}
	out := make([]string, 0, len(keys))
	weekIndex := 0
	for i := range keys {
		w := byKey[keys[i]]
		if w.dist == 0 && w.drives == 0 && w.charges == 0 {
			continue
		}
		weekIndex++
		parts := []string{fmt.Sprintf("第 %d 周", weekIndex)}
		if s := FormatDistanceDetail(w.dist); s != "" {
			parts = append(parts, s)
		}
		if w.drives > 0 {
			parts = append(parts, fmt.Sprintf("%d行程", w.drives))
		}
		if w.charges > 0 {
			parts = append(parts, fmt.Sprintf("%d充电", w.charges))
		}
		out = append(out, strings.Join(parts, " · "))
	}
	return out
}

func BuildDriveDescription(vehicleName string, d model.Drive, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	from := NormalizePlaceName(drivePointName(d, true))
	to := NormalizePlaceName(drivePointName(d, false))
	var dataLines []string
	if d.StartDate != nil && d.EndDate != nil {
		dataLines = append(dataLines, "时间："+FormatTimeRange(*d.StartDate, *d.EndDate, loc))
	} else if d.StartDate != nil {
		dataLines = append(dataLines, "时间："+d.StartDate.In(loc).Format("15:04"))
	}
	if d.DurationSeconds != nil {
		if t := FormatDurationZH(*d.DurationSeconds); t != "" {
			dataLines = append(dataLines, "耗时："+t)
		}
	} else if d.StartDate != nil && d.EndDate != nil {
		if t := FormatDurationZH(d.EndDate.Sub(*d.StartDate).Seconds()); t != "" {
			dataLines = append(dataLines, "耗时："+t)
		}
	}
	if d.Distance != nil {
		if s := FormatDistanceDetail(*d.Distance); s != "" {
			dataLines = append(dataLines, "距离："+s)
		}
	}
	if d.Consumption != nil {
		if s := FormatWhPerKm(*d.Consumption); s != "" {
			dataLines = append(dataLines, "电耗："+s)
		}
	}
	if d.StartBatteryLevel != nil && d.EndBatteryLevel != nil {
		if s := FormatBatteryRange(d.StartBatteryLevel, d.EndBatteryLevel, false); s != "" {
			dataLines = append(dataLines, "电量："+s)
		}
	}
	if d.MaxSpeed != nil && *d.MaxSpeed > 0 && !isBadFloat(*d.MaxSpeed) {
		dataLines = append(dataLines, "最高速度："+FormatSpeedKMH(*d.MaxSpeed))
	}
	var locLines []string
	locLines = append(locLines, "起点："+from)
	locLines = append(locLines, "终点："+to)
	lines := []string{vehicleName + " · 行程"}
	lines = append(lines, Section("🚗 路线", from+" → "+to)...)
	if len(dataLines) > 0 {
		lines = append(lines, Section("📊 数据", dataLines...)...)
	}
	lines = append(lines, Section("📍 位置", locLines...)...)
	if driveHasRouteRecorded(d) {
		lines = append(lines, "", "轨迹：已记录，可在 TeslaMate 看板查看")
	}
	return strings.Join(lines, "\n")
}

func BuildChargeDescription(vehicleName string, c model.Charge, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	place := strings.TrimSpace(c.Address)
	if place == "" {
		place = strings.TrimSpace(c.Geofence)
	}
	var chLines []string
	if place != "" {
		chLines = append(chLines, "地点："+NormalizePlaceName(place))
	}
	if c.StartDate != nil && c.EndDate != nil {
		chLines = append(chLines, "时间："+FormatTimeRange(*c.StartDate, *c.EndDate, loc))
	} else if c.StartDate != nil {
		chLines = append(chLines, "时间："+c.StartDate.In(loc).Format("15:04"))
	}
	var sec float64
	if c.DurationMinutes != nil {
		sec = *c.DurationMinutes * 60
	} else if c.StartDate != nil && c.EndDate != nil {
		sec = c.EndDate.Sub(*c.StartDate).Seconds()
	}
	if s := FormatDurationZH(sec); s != "" {
		chLines = append(chLines, "耗时："+s)
	}
	var dataLines []string
	if c.KwhAdded != nil && *c.KwhAdded > 0 && !isBadFloat(*c.KwhAdded) {
		dataLines = append(dataLines, "充入："+FormatKWhInDetailLine(*c.KwhAdded))
	}
	if c.StartBatteryLevel != nil && c.EndBatteryLevel != nil {
		if s := FormatBatteryRange(c.StartBatteryLevel, c.EndBatteryLevel, false); s != "" {
			dataLines = append(dataLines, "电量："+s)
		}
	}
	if c.Cost != nil && *c.Cost > 0 && !isBadFloat(*c.Cost) {
		if s := FormatCost(*c.Cost); s != "" {
			dataLines = append(dataLines, "费用："+s)
		}
	}
	if c.MaxPower != nil && *c.MaxPower > 0 && !isBadFloat(*c.MaxPower) {
		dataLines = append(dataLines, "最高功率："+FormatPowerKW(*c.MaxPower))
	}
	lines := []string{vehicleName + " · 充电"}
	lines = append(lines, Section("⚡ 充电", chLines...)...)
	if len(dataLines) > 0 {
		lines = append(lines, Section("📊 数据", dataLines...)...)
	}
	return strings.Join(lines, "\n")
}
