package calendar

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
)

func BuildDailySummaries(drives []model.Drive, charges []model.Charge, updates []model.Update, loc *time.Location) []model.DailySummary {
	sort.Slice(drives, func(i, j int) bool {
		return safeTime(drives[i].StartDate).Before(safeTime(drives[j].StartDate))
	})
	sort.Slice(charges, func(i, j int) bool {
		return safeTime(charges[i].StartDate).Before(safeTime(charges[j].StartDate))
	})
	sort.Slice(updates, func(i, j int) bool {
		return safeTime(updates[i].StartDate).Before(safeTime(updates[j].StartDate))
	})
	byDay := map[string]*model.DailySummary{}
	get := func(t time.Time) *model.DailySummary {
		k := t.In(loc).Format("2006-01-02")
		if s, ok := byDay[k]; ok {
			return s
		}
		day, _ := time.ParseInLocation("2006-01-02", k, loc)
		s := &model.DailySummary{Day: day}
		byDay[k] = s
		return s
	}
	for _, d := range drives {
		if d.StartDate == nil {
			continue
		}
		s := get(*d.StartDate)
		s.DriveCount++
		if d.Distance != nil {
			s.Distance += *d.Distance
		}
		if d.DurationSeconds != nil {
			s.DriveSeconds += *d.DurationSeconds
		}
		if d.MaxSpeed != nil && *d.MaxSpeed > s.MaxSpeed {
			s.MaxSpeed = *d.MaxSpeed
		}
		if d.Consumption != nil {
			s.Consumption += *d.Consumption
		}
		if s.StartBattery == nil && d.StartBatteryLevel != nil {
			v := *d.StartBatteryLevel
			s.StartBattery = &v
		}
		if d.EndBatteryLevel != nil {
			v := *d.EndBatteryLevel
			s.EndBattery = &v
		}
		s.DriveDetails = append(s.DriveDetails, formatDailyDriveLine(d, loc))
	}
	for _, c := range charges {
		if c.StartDate == nil {
			continue
		}
		s := get(*c.StartDate)
		s.ChargeCount++
		if c.KwhAdded != nil {
			s.KwhAdded += *c.KwhAdded
		}
		if c.Cost != nil {
			s.Cost += *c.Cost
		}
		if c.MaxPower != nil && *c.MaxPower > s.MaxChargePower {
			s.MaxChargePower = *c.MaxPower
		}
		if c.EndDate != nil {
			s.ChargeSeconds += c.EndDate.Sub(*c.StartDate).Seconds()
		}
		if s.StartBattery == nil && c.StartBatteryLevel != nil {
			v := *c.StartBatteryLevel
			s.StartBattery = &v
		}
		if c.EndBatteryLevel != nil {
			v := *c.EndBatteryLevel
			s.EndBattery = &v
		}
		s.ChargeDetails = append(s.ChargeDetails, formatDailyChargeLine(c, loc))
	}
	for _, u := range updates {
		if u.StartDate == nil {
			continue
		}
		s := get(*u.StartDate)
		s.UpdateCount++
		if u.Version != "" {
			s.UpdateVersions = append(s.UpdateVersions, u.Version)
		}
	}
	keys := make([]string, 0, len(byDay))
	for k := range byDay {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]model.DailySummary, 0, len(keys))
	for _, k := range keys {
		s := byDay[k]
		if s.DriveCount == 0 && s.ChargeCount == 0 && s.UpdateCount == 0 {
			continue
		}
		out = append(out, *s)
	}
	return out
}

func DailySummaryEvents(carID, vehicleName string, rows []model.DailySummary, loc *time.Location, view string, detail bool, includeItems bool, dashboardBaseURL, dashboardPath string) []Event {
	out := make([]Event, 0, len(rows))
	for _, d := range rows {
		summary := dailyTitle(vehicleName, d, view)
		desc := buildDailyDescription(vehicleName, d, includeItems, detail)
		start := time.Date(d.Day.Year(), d.Day.Month(), d.Day.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		desc = AppendLinksSection(desc, "", DashboardURL(dashboardBaseURL, dashboardPath, carID, start, end))
		uid := fmt.Sprintf("teslamate-calendar-car-%s-daily-%s", carID, d.Day.In(loc).Format("2006-01-02"))
		allDay := d.Day.In(loc)
		out = append(out, Event{
			UID:         uid,
			AllDay:      true,
			AllDayDate:  allDay,
			AllDayEnd:   allDay.AddDate(0, 0, 1),
			Summary:     summary,
			Description: desc,
			Location:    dailyLocation(d),
		})
	}
	return out
}

func WeeklySummaryEvents(carID, vehicleName string, rows []model.DailySummary, loc *time.Location, view string, detail bool, dashboardBaseURL, dashboardPath string) []Event {
	type weekAgg struct {
		WeekStart time.Time
		WeekEnd   time.Time
		Distance  float64
		DriveCnt  int
		ChargeCnt int
		UpdateCnt int
		Kwh       float64
		Days      []model.DailySummary
	}
	byWeek := map[string]*weekAgg{}
	for _, d := range rows {
		ws := startOfWeek(d.Day.In(loc))
		we := ws.AddDate(0, 0, 7)
		key := ws.Format("2006-01-02")
		if _, ok := byWeek[key]; !ok {
			byWeek[key] = &weekAgg{WeekStart: ws, WeekEnd: we}
		}
		w := byWeek[key]
		w.Distance += d.Distance
		w.DriveCnt += d.DriveCount
		w.ChargeCnt += d.ChargeCount
		w.UpdateCnt += d.UpdateCount
		w.Kwh += d.KwhAdded
		w.Days = append(w.Days, d)
	}
	keys := make([]string, 0, len(byWeek))
	for k := range byWeek {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Event, 0, len(keys))
	for _, k := range keys {
		w := byWeek[k]
		summary := weeklyTitle(vehicleName, w.Distance, w.DriveCnt, w.ChargeCnt, view)
		desc := buildWeeklyDescription(vehicleName, w.WeekStart, w.WeekEnd, loc, w.Distance, w.DriveCnt, w.ChargeCnt, w.UpdateCnt, w.Kwh, w.Days, detail)
		start := w.WeekStart
		end := w.WeekEnd
		out = append(out, Event{
			UID:         fmt.Sprintf("teslamate-calendar-car-%s-weekly-%s", carID, w.WeekStart.Format("2006-01-02")),
			AllDay:      true,
			AllDayDate:  w.WeekStart,
			AllDayEnd:   w.WeekEnd,
			Summary:     summary,
			Description: AppendLinksSection(desc, "", DashboardURL(dashboardBaseURL, dashboardPath, carID, start, end)),
			Location:    "Tesla",
		})
	}
	return out
}

func MonthlySummaryEvents(carID, vehicleName string, rows []model.DailySummary, loc *time.Location, view string, detail bool, dashboardBaseURL, dashboardPath string) []Event {
	type monthAgg struct {
		Start     time.Time
		End       time.Time
		Distance  float64
		DriveCnt  int
		ChargeCnt int
		UpdateCnt int
		Kwh       float64
		Days      []model.DailySummary
	}
	byMonth := map[string]*monthAgg{}
	for _, d := range rows {
		ds := d.Day.In(loc)
		ms := time.Date(ds.Year(), ds.Month(), 1, 0, 0, 0, 0, loc)
		me := ms.AddDate(0, 1, 0)
		key := ms.Format("2006-01")
		if _, ok := byMonth[key]; !ok {
			byMonth[key] = &monthAgg{Start: ms, End: me}
		}
		m := byMonth[key]
		m.Distance += d.Distance
		m.DriveCnt += d.DriveCount
		m.ChargeCnt += d.ChargeCount
		m.UpdateCnt += d.UpdateCount
		m.Kwh += d.KwhAdded
		m.Days = append(m.Days, d)
	}
	keys := make([]string, 0, len(byMonth))
	for k := range byMonth {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Event, 0, len(keys))
	for _, k := range keys {
		m := byMonth[k]
		summary := monthlyTitle(vehicleName, m.Distance, m.DriveCnt, m.ChargeCnt, view)
		desc := buildMonthlyDescription(vehicleName, m.Start, m.End, loc, m.Distance, m.DriveCnt, m.ChargeCnt, m.UpdateCnt, m.Kwh, m.Days, carID, detail)
		start := m.Start
		end := m.End
		out = append(out, Event{
			UID:         fmt.Sprintf("teslamate-calendar-car-%s-monthly-%s", carID, m.Start.Format("2006-01")),
			AllDay:      true,
			AllDayDate:  m.Start,
			AllDayEnd:   m.End,
			Summary:     summary,
			Description: AppendLinksSection(desc, "", DashboardURL(dashboardBaseURL, dashboardPath, carID, start, end)),
			Location:    "Tesla",
		})
	}
	return out
}

func formatDailyDriveLine(d model.Drive, loc *time.Location) string {
	startText := "--:--"
	endText := "--:--"
	if d.StartDate != nil {
		startText = d.StartDate.In(loc).Format("15:04")
	}
	if d.EndDate != nil {
		endText = d.EndDate.In(loc).Format("15:04")
	}
	from := drivePointName(d, true)
	to := drivePointName(d, false)
	line := fmt.Sprintf("- %s-%s %s → %s", startText, endText, from, to)
	if d.Distance != nil {
		line += fmt.Sprintf(" %.2fkm", *d.Distance)
	}
	if d.DurationSeconds != nil {
		line += fmt.Sprintf(" %.0fmin", *d.DurationSeconds/60)
	}
	if d.Consumption != nil {
		line += fmt.Sprintf(" %.0fWh/km", *d.Consumption)
	}
	if d.StartBatteryLevel != nil && d.EndBatteryLevel != nil {
		line += fmt.Sprintf(" %.0f%%→%.0f%%", *d.StartBatteryLevel, *d.EndBatteryLevel)
	}
	if m := GoogleRouteURL(d.StartLat, d.StartLng, d.EndLat, d.EndLng); m != "" {
		line += "\n地图：" + m
	}
	return line
}

func formatDailyChargeLine(c model.Charge, loc *time.Location) string {
	startText := "--:--"
	endText := "--:--"
	if c.StartDate != nil {
		startText = c.StartDate.In(loc).Format("15:04")
	}
	if c.EndDate != nil {
		endText = c.EndDate.In(loc).Format("15:04")
	}
	line := fmt.Sprintf("- %s-%s", startText, endText)
	if c.KwhAdded != nil {
		line += fmt.Sprintf(" +%.2fkWh", *c.KwhAdded)
	}
	if c.StartBatteryLevel != nil && c.EndBatteryLevel != nil {
		line += fmt.Sprintf(" %.0f%%→%.0f%%", *c.StartBatteryLevel, *c.EndBatteryLevel)
	}
	if c.Address != "" {
		line += " @" + c.Address
	} else if c.Geofence != "" {
		line += " @" + c.Geofence
	}
	if m := GooglePointURL(c.Lat, c.Lng); m != "" {
		line += "\n地图：" + m
	}
	return line
}

func dailyTitle(vehicleName string, d model.DailySummary, view string) string {
	switch view {
	case "compact":
		parts := []string{"📊"}
		if d.Distance > 0 {
			parts = append(parts, fmt.Sprintf("%.0fkm", d.Distance))
		}
		if d.DriveCount > 0 {
			parts = append(parts, fmt.Sprintf("%d🚗", d.DriveCount))
		}
		if d.ChargeCount > 0 {
			parts = append(parts, fmt.Sprintf("%d⚡", d.ChargeCount))
		}
		if d.UpdateCount > 0 {
			parts = append(parts, fmt.Sprintf("%d⬆️", d.UpdateCount))
		}
		return strings.Join(parts, " · ")
	case "detail":
		parts := []string{"📊 " + vehicleName + " 日报"}
		if d.Distance > 0 {
			parts = append(parts, fmt.Sprintf("%.1fkm", d.Distance))
		}
		if d.DriveCount > 0 {
			parts = append(parts, fmt.Sprintf("%d行程", d.DriveCount))
		}
		if d.ChargeCount > 0 {
			parts = append(parts, fmt.Sprintf("%d充电", d.ChargeCount))
		}
		if d.KwhAdded > 0 {
			parts = append(parts, fmt.Sprintf("+%.1fkWh", d.KwhAdded))
		}
		if d.UpdateCount > 0 {
			parts = append(parts, fmt.Sprintf("%d更新", d.UpdateCount))
		}
		return strings.Join(parts, " · ")
	default:
		parts := []string{"📊 " + vehicleName}
		if d.Distance > 0 {
			parts = append(parts, fmt.Sprintf("%.0fkm", d.Distance))
		}
		if d.DriveCount > 0 {
			parts = append(parts, fmt.Sprintf("%d行程", d.DriveCount))
		}
		if d.ChargeCount > 0 {
			parts = append(parts, fmt.Sprintf("%d充电", d.ChargeCount))
		}
		if d.UpdateCount > 0 {
			parts = append(parts, fmt.Sprintf("%d更新", d.UpdateCount))
		}
		return strings.Join(parts, " · ")
	}
}

func weeklyTitle(vehicleName string, distance float64, driveCnt, chargeCnt int, view string) string {
	switch view {
	case "compact":
		return fmt.Sprintf("📊 本周 · %.0fkm · %d🚗 · %d⚡", distance, driveCnt, chargeCnt)
	case "detail":
		return fmt.Sprintf("📊 %s 周报 · %.1fkm · %d行程 · %d充电", vehicleName, distance, driveCnt, chargeCnt)
	default:
		return fmt.Sprintf("📊 %s 周报 · %.0fkm · %d行程 · %d充电", vehicleName, distance, driveCnt, chargeCnt)
	}
}

func monthlyTitle(vehicleName string, distance float64, driveCnt, chargeCnt int, view string) string {
	switch view {
	case "compact":
		return fmt.Sprintf("📊 本月 · %.0fkm · %d🚗 · %d⚡", distance, driveCnt, chargeCnt)
	case "detail":
		return fmt.Sprintf("📊 %s 月报 · %.1fkm · %d行程 · %d充电", vehicleName, distance, driveCnt, chargeCnt)
	default:
		return fmt.Sprintf("📊 %s 月报 · %.0fkm · %d行程 · %d充电", vehicleName, distance, driveCnt, chargeCnt)
	}
}

func buildWeeklyDescription(vehicleName string, weekStart, weekEnd time.Time, loc *time.Location, distance float64, driveCnt, chargeCnt, updateCnt int, kwh float64, days []model.DailySummary, detail bool) string {
	lines := []string{
		"车辆：" + vehicleName,
		fmt.Sprintf("周期：%s ~ %s", weekStart.In(loc).Format("2006-01-02"), weekEnd.AddDate(0, 0, -1).In(loc).Format("2006-01-02")),
		"",
		"━━━━━━━━━━━━",
		"📊 周期摘要",
		"━━━━━━━━━━━━",
	}
	if distance > 0 {
		lines = append(lines, fmt.Sprintf("总里程：%.1f km", distance))
	}
	if driveCnt > 0 {
		lines = append(lines, fmt.Sprintf("行程：%d 次", driveCnt))
	}
	if chargeCnt > 0 {
		lines = append(lines, fmt.Sprintf("充电：%d 次", chargeCnt))
	}
	if kwh > 0 {
		lines = append(lines, fmt.Sprintf("充入：%.1f kWh", kwh))
	}
	if updateCnt > 0 {
		lines = append(lines, fmt.Sprintf("软件更新：%d 次", updateCnt))
	}
	if !detail {
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "", "━━━━━━━━━━━━", "📅 分日汇总", "━━━━━━━━━━━━")
	for _, d := range days {
		parts := []string{d.Day.In(loc).Format("01-02")}
		if d.Distance > 0 {
			parts = append(parts, fmt.Sprintf("%.1fkm", d.Distance))
		}
		if d.DriveCount > 0 {
			parts = append(parts, fmt.Sprintf("%d行程", d.DriveCount))
		}
		if d.ChargeCount > 0 {
			parts = append(parts, fmt.Sprintf("%d充电", d.ChargeCount))
		}
		if d.UpdateCount > 0 {
			parts = append(parts, fmt.Sprintf("%d更新", d.UpdateCount))
		}
		lines = append(lines, "- "+strings.Join(parts, " · "))
	}
	return strings.Join(lines, "\n")
}

func buildMonthlyDescription(vehicleName string, monthStart, monthEnd time.Time, loc *time.Location, distance float64, driveCnt, chargeCnt, updateCnt int, kwh float64, days []model.DailySummary, carID string, detail bool) string {
	lines := []string{
		"车辆：" + vehicleName,
		"月份：" + monthStart.In(loc).Format("2006-01"),
		"",
		"━━━━━━━━━━━━",
		"📊 月度摘要",
		"━━━━━━━━━━━━",
	}
	if distance > 0 {
		lines = append(lines, fmt.Sprintf("总里程：%.1f km", distance))
	}
	if driveCnt > 0 {
		lines = append(lines, fmt.Sprintf("行程：%d 次", driveCnt))
	}
	if chargeCnt > 0 {
		lines = append(lines, fmt.Sprintf("充电：%d 次", chargeCnt))
	}
	if kwh > 0 {
		lines = append(lines, fmt.Sprintf("充入：%.1f kWh", kwh))
	}
	if updateCnt > 0 {
		lines = append(lines, fmt.Sprintf("软件更新：%d 次", updateCnt))
	}
	if !detail {
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "", "━━━━━━━━━━━━", "📅 分周汇总", "━━━━━━━━━━━━")
	weekly := WeeklySummaryEvents(carID, vehicleName, days, loc, "normal", false, "", "")
	for _, w := range weekly {
		lines = append(lines, "- "+w.Summary)
	}
	return strings.Join(lines, "\n")
}

func buildDailyDescription(vehicleName string, d model.DailySummary, includeItems, detail bool) string {
	lines := []string{
		"日期：" + d.Day.Format("2006-01-02"),
		"车辆：" + vehicleName,
		"",
		"━━━━━━━━━━━━",
		"📊 当日摘要",
		"━━━━━━━━━━━━",
	}
	if d.DriveCount > 0 {
		lines = append(lines, fmt.Sprintf("行程：%d 次", d.DriveCount))
	}
	if d.Distance > 0 {
		lines = append(lines, fmt.Sprintf("里程：%.2f km", d.Distance))
	}
	if d.DriveSeconds > 0 {
		lines = append(lines, fmt.Sprintf("行驶：%.0f min", d.DriveSeconds/60))
	}
	if d.MaxSpeed > 0 {
		lines = append(lines, fmt.Sprintf("最高速度：%.0f km/h", d.MaxSpeed))
	}
	if d.ChargeCount > 0 {
		lines = append(lines, fmt.Sprintf("充电：%d 次", d.ChargeCount))
	}
	if d.KwhAdded > 0 {
		lines = append(lines, fmt.Sprintf("充入：%.2f kWh", d.KwhAdded))
	}
	if d.ChargeSeconds > 0 {
		lines = append(lines, fmt.Sprintf("充电时长：%.0f min", d.ChargeSeconds/60))
	}
	if d.Cost > 0 {
		lines = append(lines, fmt.Sprintf("费用：%.2f", d.Cost))
	}
	if d.MaxChargePower > 0 {
		lines = append(lines, fmt.Sprintf("充电峰值：%.0f kW", d.MaxChargePower))
	}
	if detail && d.StartBattery != nil && d.EndBattery != nil {
		lines = append(lines, fmt.Sprintf("电量：%.0f%% → %.0f%%", *d.StartBattery, *d.EndBattery))
	}
	if includeItems && len(d.DriveDetails) > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "🚗 行程明细", "━━━━━━━━━━━━")
		lines = append(lines, d.DriveDetails...)
	}
	if includeItems && len(d.ChargeDetails) > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "⚡ 充电明细", "━━━━━━━━━━━━")
		lines = append(lines, d.ChargeDetails...)
	}
	if len(d.UpdateVersions) > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "⬆️ 软件更新", "━━━━━━━━━━━━")
		for _, v := range d.UpdateVersions {
			lines = append(lines, "版本："+v)
		}
	} else if d.UpdateCount > 0 {
		lines = append(lines, "", "━━━━━━━━━━━━", "⬆️ 软件更新", "━━━━━━━━━━━━")
		lines = append(lines, fmt.Sprintf("次数：%d 次", d.UpdateCount))
	}
	return strings.Join(lines, "\n")
}

func dailyLocation(d model.DailySummary) string {
	if d.DriveCount > 0 && d.ChargeCount > 0 {
		return fmt.Sprintf("Tesla · %d行程 · %d充电", d.DriveCount, d.ChargeCount)
	}
	if d.DriveCount > 0 {
		return fmt.Sprintf("Tesla · %d行程", d.DriveCount)
	}
	if d.ChargeCount > 0 {
		return fmt.Sprintf("Tesla · %d充电", d.ChargeCount)
	}
	return "Tesla"
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func startOfWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-(wd-1), 0, 0, 0, 0, t.Location())
}
