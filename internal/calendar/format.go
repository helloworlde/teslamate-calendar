package calendar

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func isBadFloat(f float64) bool {
	return math.IsNaN(f) || math.IsInf(f, 0)
}

func FormatDurationZH(seconds float64) string {
	if isBadFloat(seconds) || seconds <= 0 {
		return ""
	}
	m := int(math.Round(seconds / 60))
	if m < 1 {
		return ""
	}
	if m < 60 {
		return fmt.Sprintf("%d分钟", m)
	}
	h := m / 60
	r := m % 60
	if r == 0 {
		return fmt.Sprintf("%d小时", h)
	}
	return fmt.Sprintf("%d小时%d分钟", h, r)
}

func FormatDistanceTitle(km float64) string {
	if isBadFloat(km) || km <= 0 {
		return ""
	}
	return fmt.Sprintf("%.0fkm", math.Round(km))
}

func FormatDistanceTitleOneDecimal(km float64) string {
	if isBadFloat(km) || km <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1fkm", km)
}

func FormatDistanceDetail(km float64) string {
	if isBadFloat(km) || km <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1f km", round1f(km))
}

func round1f(v float64) float64 {
	return math.Round(v*10) / 10
}

func FormatKWhTitle(kwh float64) string {
	if isBadFloat(kwh) || kwh <= 0 {
		return ""
	}
	return fmt.Sprintf("+%.1fkWh", kwh)
}

func FormatKWhDetail(kwh float64) string {
	if isBadFloat(kwh) || kwh <= 0 {
		return ""
	}
	return fmt.Sprintf("+%.1f kWh", kwh)
}

func FormatKWhInDetailLine(kwh float64) string {
	if isBadFloat(kwh) || kwh <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1f kWh", kwh)
}

func FormatBatteryRange(start, end *float64, spaced bool) string {
	if start == nil || end == nil {
		return ""
	}
	if isBadFloat(*start) || isBadFloat(*end) {
		return ""
	}
	if spaced {
		return fmt.Sprintf("%.0f%% → %.0f%%", *start, *end)
	}
	return fmt.Sprintf("%.0f%%→%.0f%%", *start, *end)
}

func FormatTimeRange(start, end time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	s := start.In(loc)
	e := end.In(loc)
	if !e.After(s) {
		return s.Format("15:04")
	}
	if s.Year() == e.Year() && s.YearDay() == e.YearDay() {
		return fmt.Sprintf("%s-%s", s.Format("15:04"), e.Format("15:04"))
	}
	return fmt.Sprintf("%s-次日 %s", s.Format("15:04"), e.Format("15:04"))
}

func FormatCost(c float64) string {
	if isBadFloat(c) || c <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", c)
}

func FormatSpeedKMH(v float64) string {
	if isBadFloat(v) || v <= 0 {
		return ""
	}
	if math.Round(v) == v {
		return fmt.Sprintf("%.0f km/h", v)
	}
	return fmt.Sprintf("%.0f km/h", v)
}

func FormatPowerKW(v float64) string {
	if isBadFloat(v) || v <= 0 {
		return ""
	}
	if math.Round(v) == v {
		return fmt.Sprintf("%.0f kW", v)
	}
	return fmt.Sprintf("%.0f kW", v)
}

func FormatWhPerKm(v float64) string {
	if isBadFloat(v) || v < 0 {
		return ""
	}
	return fmt.Sprintf("%.0f Wh/km", v)
}

func NormalizePlaceName(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return s
	}
	if strings.Count(s, ",") != 1 {
		return s
	}
	idx := strings.Index(s, ",")
	a := strings.TrimSpace(s[:idx])
	b := strings.TrimSpace(s[idx+1:])
	if a == "" || b == "" {
		return s
	}
	return a + "（" + b + "）"
}
