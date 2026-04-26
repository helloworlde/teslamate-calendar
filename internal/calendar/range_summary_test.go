package calendar

import (
	"testing"
	"time"

	"teslamate-calendar/internal/model"
)

func TestRangeSummariesDayWeekMonth(t *testing.T) {
	loc := time.Local
	s1 := time.Date(2026, 4, 1, 9, 0, 0, 0, loc)
	e1 := s1.Add(30 * time.Minute)
	s2 := time.Date(2026, 4, 8, 9, 0, 0, 0, loc)
	e2 := s2.Add(30 * time.Minute)
	d1, d2 := 10.0, 20.0

	rows := BuildDailySummaries([]model.Drive{
		{StartDate: &s1, EndDate: &e1, Distance: &d1},
		{StartDate: &s2, EndDate: &e2, Distance: &d2},
	}, nil, nil, loc)
	dayEvents := DailySummaryEvents("1", "Model 3", rows, loc, "normal", true, true, "", "")
	weekEvents := WeeklySummaryEvents("1", "Model 3", rows, loc, "normal", true, "", "")
	monthEvents := MonthlySummaryEvents("1", "Model 3", rows, loc, "normal", true, "", "")
	if len(dayEvents) != 2 {
		t.Fatalf("day range unexpected events: %d", len(dayEvents))
	}
	if len(weekEvents) == 0 {
		t.Fatal("week range should have events")
	}
	if len(monthEvents) == 0 {
		t.Fatal("month range should have events")
	}
}
