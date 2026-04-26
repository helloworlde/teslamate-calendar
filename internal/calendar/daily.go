package calendar

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/model"
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
		s.DriveDetails = append(s.DriveDetails, formatDailyDriveBlock(d, loc))
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
		s.ChargeDetails = append(s.ChargeDetails, formatDailyChargeBlock(c, loc))
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

func DailySummaryEvents(carID, vehicleName string, rows []model.DailySummary, loc *time.Location, view string, detail bool, dashboardTmpl string) []Event {
	out := make([]Event, 0, len(rows))
	for _, d := range rows {
		summary := dailyTitle(vehicleName, d)
		desc := BuildDailyDescription(vehicleName, d, detail, loc)
		start := time.Date(d.Day.Year(), d.Day.Month(), d.Day.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		dash := RenderDashboardURL(dashboardTmpl, timesToDashboardArgs(carID, "day", "", start, end))
		desc = AppendLinksSection(desc, "", dash)
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

func WeeklySummaryEvents(carID, vehicleName string, rows []model.DailySummary, loc *time.Location, view string, detail bool, dashboardTmpl string) []Event {
	type weekAgg struct {
		WeekStart     time.Time
		WeekEnd       time.Time
		Distance      float64
		DriveCnt      int
		ChargeCnt     int
		UpdateCnt     int
		Kwh           float64
		Days          []model.DailySummary
		TotalDriveSec float64
		Cost          float64
		MaxSpeed      float64
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
		w.TotalDriveSec += d.DriveSeconds
		w.Cost += d.Cost
		if d.MaxSpeed > w.MaxSpeed {
			w.MaxSpeed = d.MaxSpeed
		}
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
		summary := weeklyTitle(vehicleName, w.DriveCnt, w.ChargeCnt, w.UpdateCnt, w.Distance, w.Kwh)
		desc := BuildWeeklyDescription(vehicleName, w.WeekStart, w.WeekEnd, loc, w.Distance, w.DriveCnt, w.ChargeCnt, w.UpdateCnt, w.Kwh, w.MaxSpeed, w.Cost, w.TotalDriveSec, w.Days, detail)
		start := w.WeekStart
		end := w.WeekEnd
		dash := RenderDashboardURL(dashboardTmpl, timesToDashboardArgs(carID, "week", "", start, end))
		out = append(out, Event{
			UID:         fmt.Sprintf("teslamate-calendar-car-%s-weekly-%s", carID, w.WeekStart.Format("2006-01-02")),
			AllDay:      true,
			AllDayDate:  w.WeekStart,
			AllDayEnd:   w.WeekEnd,
			Summary:     summary,
			Description: AppendLinksSection(desc, "", dash),
			Location:    "Tesla",
		})
	}
	return out
}

func MonthlySummaryEvents(carID, vehicleName string, rows []model.DailySummary, loc *time.Location, view string, detail bool, dashboardTmpl string) []Event {
	type monthAgg struct {
		Start         time.Time
		End           time.Time
		Distance      float64
		DriveCnt      int
		ChargeCnt     int
		UpdateCnt     int
		Kwh           float64
		Days          []model.DailySummary
		TotalDriveSec float64
		Cost          float64
		MaxSpeed      float64
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
		m.TotalDriveSec += d.DriveSeconds
		m.Cost += d.Cost
		if d.MaxSpeed > m.MaxSpeed {
			m.MaxSpeed = d.MaxSpeed
		}
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
		summary := monthlyTitle(vehicleName, m.DriveCnt, m.ChargeCnt, m.UpdateCnt, m.Distance, m.Kwh)
		desc := BuildMonthlyDescription(vehicleName, m.Start, m.End, loc, m.Distance, m.DriveCnt, m.ChargeCnt, m.UpdateCnt, m.Kwh, m.MaxSpeed, m.Cost, m.TotalDriveSec, m.Days, carID, detail)
		start := m.Start
		end := m.End
		dash := RenderDashboardURL(dashboardTmpl, timesToDashboardArgs(carID, "month", "", start, end))
		out = append(out, Event{
			UID:         fmt.Sprintf("teslamate-calendar-car-%s-monthly-%s", carID, m.Start.Format("2006-01")),
			AllDay:      true,
			AllDayDate:  m.Start,
			AllDayEnd:   m.End,
			Summary:     summary,
			Description: AppendLinksSection(desc, "", dash),
			Location:    "Tesla",
		})
	}
	return out
}

func summarySegments(driveCnt, chargeCnt, updateCnt int) string {
	var parts []string
	if driveCnt > 0 {
		parts = append(parts, fmt.Sprintf("🚗*%d", driveCnt))
	}
	if chargeCnt > 0 {
		parts = append(parts, fmt.Sprintf("⚡️*%d", chargeCnt))
	}
	if updateCnt > 0 {
		parts = append(parts, fmt.Sprintf("⬆️*%d", updateCnt))
	}
	return strings.Join(parts, "/")
}

func buildSummaryTitle(driveCnt, chargeCnt, updateCnt int, vehicleName string, distance, kwhAdded float64) string {
	v := strings.TrimSpace(vehicleName)
	if v == "" {
		v = "Tesla"
	}
	core := summarySegments(driveCnt, chargeCnt, updateCnt)
	mid := ""
	onlyDrive := driveCnt > 0 && chargeCnt == 0 && updateCnt == 0
	onlyCharge := chargeCnt > 0 && driveCnt == 0 && updateCnt == 0
	if onlyDrive {
		if km := FormatDistanceTitle(distance); km != "" {
			mid = "/" + km
		}
	} else if onlyCharge {
		if k := FormatKWhTitle(kwhAdded); k != "" {
			mid = "/" + k
		}
	}
	if core == "" {
		if mid != "" {
			return strings.TrimPrefix(mid, "/") + " - " + v
		}
		return v
	}
	return core + mid + " - " + v
}

func dailyTitle(vehicleName string, d model.DailySummary) string {
	return buildSummaryTitle(d.DriveCount, d.ChargeCount, d.UpdateCount, vehicleName, d.Distance, d.KwhAdded)
}

func weeklyTitle(vehicleName string, driveCnt, chargeCnt, updateCnt int, distance, kwhAdded float64) string {
	return buildSummaryTitle(driveCnt, chargeCnt, updateCnt, vehicleName, distance, kwhAdded)
}

func monthlyTitle(vehicleName string, driveCnt, chargeCnt, updateCnt int, distance, kwhAdded float64) string {
	return buildSummaryTitle(driveCnt, chargeCnt, updateCnt, vehicleName, distance, kwhAdded)
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
