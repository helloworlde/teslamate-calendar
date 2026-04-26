package calendar

import (
	"strings"
	"testing"
	"time"

	"teslamate-calendar/internal/model"
)

func TestDailyDescriptionHideZeroChargeFields(t *testing.T) {
	day := time.Date(2026, 4, 24, 0, 0, 0, 0, time.Local)
	rows := []model.DailySummary{{
		Day:          day,
		DriveCount:   2,
		Distance:     16.2,
		DriveSeconds: 2400,
	}}
	events := DailySummaryEvents("1", "Model 3", rows, time.Local, "normal", true, true, "", "")
	if len(events) != 1 {
		t.Fatalf("unexpected events: %d", len(events))
	}
	desc := events[0].Description
	if strings.Contains(desc, "充电：0") || strings.Contains(desc, "费用：0.00") || strings.Contains(desc, "最大功率：0") {
		t.Fatalf("zero-value charge fields should be hidden: %s", desc)
	}
}

func TestDailyDescriptionStructure(t *testing.T) {
	day := time.Date(2026, 4, 24, 0, 0, 0, 0, time.Local)
	rows := []model.DailySummary{{
		Day:           day,
		DriveCount:    1,
		Distance:      10,
		DriveSeconds:  1200,
		DriveDetails:  []string{"09:33-10:00\nA → B\n9.78 km · 26 min"},
		ChargeCount:   1,
		KwhAdded:      20,
		ChargeDetails: []string{"21:11-07:38\n+20.0 kWh · 16%→100% · 627 min"},
	}}
	events := DailySummaryEvents("1", "Model 3", rows, time.Local, "detail", true, true, "", "")
	desc := events[0].Description
	for _, part := range []string{"━━━━━━━━━━━━", "📊 当日摘要", "🚗 行程明细", "⚡ 充电明细"} {
		if !strings.Contains(desc, part) {
			t.Fatalf("missing section %s", part)
		}
	}
}

func TestWeeklyDescriptionSections(t *testing.T) {
	loc := time.Local
	day := time.Date(2026, 4, 24, 0, 0, 0, 0, loc)
	rows := []model.DailySummary{{
		Day:         day,
		DriveCount:  1,
		Distance:    5,
		ChargeCount: 0,
	}}
	weeks := WeeklySummaryEvents("1", "Model 3", rows, loc, "normal", true, "", "")
	if len(weeks) != 1 {
		t.Fatalf("expected 1 week, got %d", len(weeks))
	}
	d := weeks[0].Description
	if !strings.Contains(d, "📊 周期摘要") || !strings.Contains(d, "📅 分日汇总") {
		t.Fatalf("missing weekly sections: %s", d)
	}
	compact := WeeklySummaryEvents("1", "Model 3", rows, loc, "normal", false, "", "")
	if len(compact) != 1 {
		t.Fatal("expected 1 week when detail false")
	}
	if strings.Contains(compact[0].Description, "分日汇总") {
		t.Fatalf("detail false should not include 分日汇总")
	}
}

func TestDailyDetailsSortedByStartTime(t *testing.T) {
	loc := time.Local
	s1 := time.Date(2026, 4, 24, 20, 5, 0, 0, loc)
	e1 := time.Date(2026, 4, 24, 20, 19, 0, 0, loc)
	s2 := time.Date(2026, 4, 24, 9, 33, 0, 0, loc)
	e2 := time.Date(2026, 4, 24, 10, 0, 0, 0, loc)
	d1 := 6.09
	d2 := 9.78
	rows := BuildDailySummaries([]model.Drive{
		{StartDate: &s1, EndDate: &e1, Distance: &d1},
		{StartDate: &s2, EndDate: &e2, Distance: &d2},
	}, nil, nil, loc)
	if len(rows) != 1 || len(rows[0].DriveDetails) != 2 {
		t.Fatalf("unexpected summary rows")
	}
	if !strings.Contains(rows[0].DriveDetails[0], "09:33") {
		t.Fatalf("details not sorted ascending: %+v", rows[0].DriveDetails)
	}
}
